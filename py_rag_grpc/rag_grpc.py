# -*- coding: utf-8 -*-
import os
from concurrent import futures
import grpc
from dotenv import load_dotenv

# 1. 核心 LangChain 组件（删除了多余不用的组件，保留了核心与大模型组件）
from langchain_text_splitters import RecursiveCharacterTextSplitter
from langchain_community.embeddings import DashScopeEmbeddings
from langchain_chroma import Chroma
from langchain_openai import ChatOpenAI

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


class RagServiceServicer(rag_pb2_grpc.RagServiceServicer):

    def SearchRag(self, request, context):
        try:
            query=request.query
            results = vector_store.similarity_search(query=query, k=5)
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


