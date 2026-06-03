# -*- coding: utf-8 -*-
import os
from concurrent import futures
import grpc
from dotenv import load_dotenv
from langchain.agents.middleware import wrap_tool_call
from langchain_classic.retrievers import ContextualCompressionRetriever
from langchain_core.messages import ToolMessage, AIMessage

# 1. 核心 LangChain 组件（删除了多余不用的组件，保留了核心与大模型组件）
from langchain_text_splitters import RecursiveCharacterTextSplitter
from langchain_community.embeddings import DashScopeEmbeddings
from langchain_chroma import Chroma
from langchain_openai import ChatOpenAI
from  langchain_core.retrievers import *
from langchain_community.document_compressors import FlashrankRerank
from langchain.agents import create_agent, middleware, AgentState
from langgraph.prebuilt.tool_node import ToolCallRequest
from langchain_core.tools import tool

# 2. 引入生成的 gRPC 代码
from pb import rag_pb2_grpc, rag_pb2

# 加载配置
load_dotenv()
EMBED_API_KEY = os.getenv("EMBED_API_KEY")
apiKey = os.getenv("LLM_API_KEY")
base_url = os.getenv("LLM_BASE_URL")

# 预留给未来 RAG 检索/流式对话使用的 LLM 对象
openai = ChatOpenAI(
    api_key=apiKey,
    base_url=base_url,
    model="mimo-v2.5-pro",
    temperature=0,
)

# 统一维护的本地 Chroma 向量库实例
vector_store = Chroma(
    collection_name="test",
    persist_directory="./chroma_db",
    embedding_function=DashScopeEmbeddings(dashscope_api_key=EMBED_API_KEY),
    collection_metadata={"hnsw:space": "cosine"}
)

base_retriever = vector_store.as_retriever(search_kwargs={"k": 10})
# 初始化本地轻量级重排器，设置最终只返回最相关的 5 个结果
reranker = FlashrankRerank(top_n=2)

compression_retriever = ContextualCompressionRetriever(
    base_compressor=reranker,
    base_retriever=base_retriever
)

@wrap_tool_call
def log_tool_call(request: ToolCallRequest, handler):
    """在工具调用前后添加日志"""
    tool_name = request.tool_call.get("name")
    args = request.tool_call.get("args", {})

    print(f"\n🔧 开始调用工具: {tool_name}")
    print(f"📝 参数: {args}")

    # 执行实际的工具调用
    result = handler(request)

    print(f"✅ 工具 {tool_name} 执行完成")
    print(f"📤 返回结果长度: {len(str(result))} 字符")
    return result


@tool(description="获取资料库中的相关资料，传入str关键词，返回list[Any]相关资料")
def SearchRag(keywords:str)-> list[Any] | None:
    try:
        query =keywords
        results = compression_retriever.invoke(query)
        items_list = []
        for doc in results:
            score = doc.metadata.get("relevance_score", 0.0)
            if score > 0.5:
                # doc 直接就是 Document 对象，通过 .page_content 拿到文本
                items_list.append(doc.page_content)
        return items_list
    except Exception as e:
        print(f"[错误] 搜索失败: {str(e)}")
        return None

agent=create_agent(
    model=ChatOpenAI(
        api_key=apiKey,
        base_url=base_url,
        model="mimo-v2.5",
        temperature =0,
    ),
    tools=[SearchRag],
    system_prompt="你是一个智能AI聊天机器人，可以调用工具搜索最多3次来回答用户问题,不要返回特殊格式的内容，直接把搜索到的相关资料用自然语言回答用户",
    middleware=[log_tool_call],
)

class RagServiceServicer(rag_pb2_grpc.RagServiceServicer):

    def SearchRag(self, request, context):
        try:
            query=request.query
            results = compression_retriever.invoke(query)
            items_list = []
            for result in results:
                item=rag_pb2.RagSearchResponse.RagItem(
                    article_id=result.metadata.get("article_id", ""),
                    title = result.metadata.get("title", ""),
                )
                items_list.append(item)
            return rag_pb2.RagSearchResponse(
                items=items_list
            )
        except Exception as e:
            print(f"[错误] 搜索失败: {str(e)}")
            # 告知 Go 端：内部发生崩溃
            return rag_pb2.RagSearchResponse()


    def AddRag(self, request, context):
        try:
            # 接收 Go 端传过来的参数
            title = request.title
            article_id = request.article_id
            content = request.content

            print(f"==================================================")
            print(f"[Go端请求] 收到文章同步 -> ID: {article_id}, 标题: {title}")

            # 3. 文本切片
            text_splitter = RecursiveCharacterTextSplitter(
                chunk_size=600,
                chunk_overlap=60,
                length_function=len
            )
            chunks = text_splitter.split_text(content)
            print(f"[切片成功] 共生成 {len(chunks)} 个文本段落碎片")

            # 4. 动态生成向量库所需要的数据结构
            metadatas = []
            ids = []

            for i in range(len(chunks)):
                metadatas.append({
                    "article_id": article_id,
                    "title": title,
                    "chunk_index": i
                })
                # 使用 "文章ID_碎片索引" 作为向量库的唯一 ID，防止脏数据
                ids.append(f"{article_id}_{i}")

            # 5. 轰入本地 Chroma 数据库
            vector_store.add_texts(
                texts=chunks,
                metadatas=metadatas,
                ids=ids
            )

            print(f"[建索成功] 文章 ID: {article_id} 已成功存入 Chroma 向量库 (test集合)")
            print(f"==================================================")

            # 告知 Go 端：同步成功
            return rag_pb2.RagResponse(ok=True)

        except Exception as e:
            print(f"[严重错误] 向量化同步失败: {str(e)}")
            # 告知 Go 端：内部发生崩溃
            return rag_pb2.RagResponse(ok=False)

    def AskRag(self, request, context):
        response_dict=agent.invoke({"messages": [{"role": "user", "content": request.question}]})
        message=response_dict.get("messages", [])
        return rag_pb2.RagQuestionResponse(
            answer=message[-1].content
        )


def serve():
    # 修正了原本极其严重的缩进错误
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    rag_pb2_grpc.add_RagServiceServicer_to_server(RagServiceServicer(), server)
    server.add_insecure_port('[::]:50051')
    print("🚀 Python mimo-RAG 微服务已成功启动，正在监听 50051 端口...")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()


