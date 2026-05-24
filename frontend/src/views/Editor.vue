<template>
  <div class="editor-page">
    <div class="main-container">
      <div class="editor-header">
        <h1 class="page-title">
          <el-icon><EditPen /></el-icon>
          {{ isEditing ? '编辑文章' : '写新文章' }}
        </h1>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="editor-form"
        @submit.prevent
      >
        <!-- Title -->
        <el-form-item prop="title">
          <el-input
            v-model="form.title"
            placeholder="请输入文章标题"
            size="large"
            class="title-input"
            maxlength="120"
            show-word-limit
          />
        </el-form-item>

        <!-- Cover Image Upload -->
        <el-form-item label="封面图" prop="cover_img">
          <div class="cover-upload">
            <el-upload
              class="cover-uploader"
              :show-file-list="false"
              :before-upload="beforeCoverUpload"
              :http-request="handleCoverUpload"
              accept="image/*"
            >
              <div v-if="form.cover_img" class="cover-preview">
                <el-image
                  :src="getImageUrl(form.cover_img)"
                  fit="cover"
                  class="preview-img"
                />
                <div class="cover-overlay">
                  <el-icon size="24"><RefreshRight /></el-icon>
                  <span>更换封面</span>
                </div>
                <el-button
                  class="remove-cover"
                  type="danger"
                  :icon="Delete"
                  circle
                  size="small"
                  @click.stop="form.cover_img = ''"
                />
              </div>
              <div v-else class="cover-placeholder">
                <el-icon size="32"><Plus /></el-icon>
                <span>上传封面图</span>
                <span class="hint">建议尺寸 16:9</span>
              </div>
            </el-upload>
          </div>
        </el-form-item>

        <!-- Tags -->
        <el-form-item label="标签" prop="tag_list">
          <div class="tags-section">
            <div class="selected-tags">
              <el-tag
                v-for="tag in form.tag_list"
                :key="tag.id"
                closable
                type="primary"
                effect="plain"
                @close="removeTag(tag)"
              >
                {{ tag.name }}
              </el-tag>
            </div>
            <div class="tag-input-wrapper">
              <el-select
                v-model="selectedTagToAdd"
                placeholder="选择已有标签添加"
                filterable
                clearable
                size="default"
                class="tag-select"
                @change="onSelectExistingTag"
              >
                <el-option
                  v-for="tag in availableTags"
                  :key="tag.id"
                  :label="tag.name"
                  :value="tag.id"
                />
              </el-select>
              <el-input
                v-model="newTagName"
                placeholder="输入新标签名"
                size="default"
                class="tag-create-input"
                clearable
                @keyup.enter="handleCreateTag"
              >
                <template #append>
                  <el-button
                    :icon="Plus"
                    :loading="creatingTag"
                    @click="handleCreateTag"
                  />
                </template>
              </el-input>
            </div>
            <!-- Suggested Tags -->
            <div v-if="suggestedTags.length > 0" class="suggested-tags">
              <span class="suggested-label">推荐标签：</span>
              <el-tag
                v-for="tag in suggestedTags"
                :key="tag.id"
                type="info"
                effect="plain"
                class="suggested-item"
                @click="addSuggestedTag(tag)"
              >
                + {{ tag.name }}
              </el-tag>
            </div>
          </div>
        </el-form-item>

        <!-- Category -->
        <el-form-item label="分类" prop="category_id">
          <el-select
            v-model="form.category_id"
            placeholder="请选择文章分类"
            filterable
            clearable
            size="default"
            style="max-width: 400px"
          >
            <el-option
              v-for="cat in categories"
              :key="cat.id"
              :label="cat.name"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>

        <!-- Content Editor -->
        <el-form-item label="文章内容" prop="content">
          <div class="content-editor">
            <div class="editor-toolbar">
              <el-button-group>
                <el-tooltip content="加粗" placement="top">
                  <el-button size="small" @click="insertFormat('**', '**')">
                    <strong>B</strong>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="斜体" placement="top">
                  <el-button size="small" @click="insertFormat('*', '*')">
                    <em>I</em>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="标题" placement="top">
                  <el-button size="small" @click="insertFormat('## ', '')">
                    H
                  </el-button>
                </el-tooltip>
                <el-tooltip content="链接" placement="top">
                  <el-button size="small" @click="insertFormat('[', '](url)')">
                    <el-icon><Link /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="代码块" placement="top">
                  <el-button size="small" @click="insertFormat('```\n', '\n```')">
                    <el-icon><Document /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="引用" placement="top">
                  <el-button size="small" @click="insertFormat('> ', '')">
                    <el-icon><ChatDotSquare /></el-icon>
                  </el-button>
                </el-tooltip>
              </el-button-group>
              <div class="toolbar-right">
                <el-switch
                  v-model="previewMode"
                  active-text="预览"
                  inactive-text="编辑"
                  inline-prompt
                  size="small"
                />
              </div>
            </div>
            <div class="editor-body">
              <el-input
                v-show="!previewMode"
                ref="editorRef"
                v-model="form.content"
                type="textarea"
                :autosize="{ minRows: 18 }"
                placeholder="开始写作吧...&#10;&#10;支持 Markdown 语法"
                class="editor-textarea"
              />
              <div
                v-show="previewMode"
                class="editor-preview"
                v-html="previewContent"
              ></div>
            </div>
          </div>
        </el-form-item>

        <!-- Submit Actions -->
        <el-form-item class="submit-actions">
          <div class="actions-wrapper">
            <el-button
              size="large"
              @click="saveDraft"
              :loading="saving"
            >
              <el-icon><FolderOpened /></el-icon>
              存为草稿
            </el-button>
            <el-button
              type="primary"
              size="large"
              @click="publishArticle"
              :loading="saving"
            >
              <el-icon><Promotion /></el-icon>
              {{ isEditing ? '更新文章' : '发布文章' }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useArticleStore } from '@/stores/article'
import { tagsApi } from '@/api/tags'
import { uploadApi } from '@/api/upload'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
import type { Tag, Category } from '@/types/common'
import {
  EditPen,
  Delete,
  Plus,
  RefreshRight,
  Link,
  Document,
  ChatDotSquare,
  FolderOpened,
  Promotion,
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const store = useArticleStore()

const formRef = ref<FormInstance>()
const editorRef = ref<any>(null)
const saving = ref(false)
const previewMode = ref(false)
const selectedTagToAdd = ref<number | null>(null)
const newTagName = ref('')
const creatingTag = ref(false)
const suggestedTags = ref<Tag[]>([])
const allTags = ref<Tag[]>([])
const categories = ref<Category[]>([])

const articleId = computed(() => String(route.params.id ?? route.params.slug ?? ''))
const isEditing = computed(() => !!articleId.value)

/** 可选择的标签（排除已选中的） */
const availableTags = computed(() => {
  const selectedIds = new Set(form.value.tag_list.map(t => t.id))
  return allTags.value.filter(t => !selectedIds.has(t.id))
})

const form = ref({
  title: '',
  content: '',
  cover_img: '',
  tag_list: [] as Tag[],
  preview: '',
  category_id: 0,
})

const rules: FormRules = {
  title: [
    { required: true, message: '请输入文章标题', trigger: 'blur' },
    { min: 2, max: 120, message: '标题长度为 2-120 个字符', trigger: 'blur' },
  ],
  content: [
    { required: true, message: '请输入文章内容', trigger: 'blur' },
    { min: 10, message: '文章内容至少 10 个字符', trigger: 'blur' },
  ],
}

/** 拼接图片完整 URL */
const getImageUrl = (path: string): string => {
  if (!path || path.startsWith('http')) return path
  return path
}

/** 生成预览摘要 */
const generatePreview = (content: string): string => {
  const plain = content
    .replace(/#{1,6}\s/g, '')
    .replace(/\*{1,2}([^*]+)\*{1,2}/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/`{1,3}[^`]*`{1,3}/g, '')
    .replace(/\n/g, ' ')
    .trim()
  const fallback = content.trim().replace(/\s+/g, ' ')
  const source = plain || fallback
  return source.substring(0, 200) + (source.length > 200 ? '...' : '')
}

const getObjectId = (obj: any): number => {
  const id = Number(obj?.id ?? obj?.ID ?? 0)
  return Number.isFinite(id) ? id : 0
}

/** Markdown 简单预览渲染 */
const previewContent = computed(() => {
  if (!form.value.content) return '<p style="color:#999">暂无内容</p>'
  return form.value.content
    .replace(/^### (.+)$/gm, '<h3>$1</h3>')
    .replace(/^## (.+)$/gm, '<h2>$1</h2>')
    .replace(/^# (.+)$/gm, '<h1>$1</h1>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/`{3}([\s\S]*?)`{3}/g, '<pre><code>$1</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/^> (.+)$/gm, '<blockquote>$1</blockquote>')
    .replace(/\[(.+?)\]\((.+?)\)/g, '<a href="$2" target="_blank">$1</a>')
    .replace(/\n\n/g, '</p><p>')
    .replace(/\n/g, '<br>')
    .replace(/^/, '<p>')
    .replace(/$/, '</p>')
})

/** 从下拉选择已有标签 */
const onSelectExistingTag = (tagId: number | null) => {
  if (tagId === null) return
  if (form.value.tag_list.length >= 5) {
    ElMessage.warning('最多添加 5 个标签')
    selectedTagToAdd.value = null
    return
  }
  const tag = allTags.value.find(t => t.id === tagId)
  if (tag && !form.value.tag_list.find(t => t.id === tag.id)) {
    form.value.tag_list.push(tag)
  }
  selectedTagToAdd.value = null
}

/** 添加推荐标签 */
const addSuggestedTag = (tag: Tag) => {
  if (form.value.tag_list.find(t => t.id === tag.id)) return
  if (form.value.tag_list.length >= 5) {
    ElMessage.warning('最多添加 5 个标签')
    return
  }
  form.value.tag_list.push(tag)
  suggestedTags.value = suggestedTags.value.filter(t => t.id !== tag.id)
}

/** 创建新标签 */
const handleCreateTag = async () => {
  const name = newTagName.value.trim()
  if (!name) return
  if (form.value.tag_list.length >= 5) {
    ElMessage.warning('最多添加 5 个标签')
    return
  }
  if (form.value.tag_list.find(t => t.name === name)) {
    ElMessage.warning('该标签已添加')
    newTagName.value = ''
    return
  }
  creatingTag.value = true
  try {
    const res = await tagsApi.createTag(name)
    ElMessage.success('标签创建成功')
    newTagName.value = ''

    // 🔥 核心：创建成功后重新拉取全量标签，保证 Single Source of Truth
    await fetchAllTags()

    // 如果后端返回了新标签的完整对象（含 id），自动选中它
    if (res && res.id) {
      const created = allTags.value.find(t => t.id === res.id)
      if (created && !form.value.tag_list.find(t => t.id === created.id)) {
        form.value.tag_list.push(created)
      }
    } else {
      // 兜底：按 name 匹配
      const match = allTags.value.find(t => t.name === name)
      if (match && !form.value.tag_list.find(t => t.id === match.id)) {
        form.value.tag_list.push(match)
      }
    }
  } catch (error: any) {
    if (error?.response?.status === 409 || error?.response?.data?.error?.includes('exist')) {
      // 标签已存在，重新拉取列表后自动选中
      await fetchAllTags()
      const existing = allTags.value.find(t => t.name === name)
      if (existing && !form.value.tag_list.find(t => t.id === existing.id)) {
        form.value.tag_list.push(existing)
      }
      newTagName.value = ''
    } else {
      ElMessage.error('创建标签失败')
    }
  } finally {
    creatingTag.value = false
  }
}

/** 移除标签 */
const removeTag = (tag: Tag) => {
  form.value.tag_list = form.value.tag_list.filter(t => t.id !== tag.id)
}

/** 封面图上传前校验 */
const beforeCoverUpload = (file: File) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5

  if (!isImage) {
    ElMessage.error('只能上传图片文件')
    return false
  }
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过 5MB')
    return false
  }
  return true
}

/** 上传封面图 */
const handleCoverUpload = async (options: UploadRequestOptions) => {
  try {
    const res = await uploadApi.uploadImages([options.file])
    form.value.cover_img = res[0] || ''
    ElMessage.success('封面上传成功')
  } catch (error) {
    ElMessage.error('封面上传失败')
  }
}

/** 插入格式化文本 */
const insertFormat = (prefix: string, suffix: string) => {
  const textarea = editorRef.value?.$el?.querySelector('textarea')
  if (!textarea) return
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const selected = form.value.content.substring(start, end)
  const replacement = `${prefix}${selected || '文本'}${suffix}`
  form.value.content =
    form.value.content.substring(0, start) +
    replacement +
    form.value.content.substring(end)
  nextTick(() => {
    textarea.focus()
    const newPos = start + prefix.length
    textarea.setSelectionRange(newPos, newPos + (selected || '文本').length)
  })
}

/** 获取文章详情（编辑模式） */
const fetchArticleForEdit = async () => {
  const id = Number(articleId.value)
  if (isNaN(id)) return
  try {
    await store.fetchArticle(String(id))
    const article = store.currentArticle
    if (article) {
      form.value.title = article.title
      form.value.content = article.content || ''
      form.value.cover_img = (article as any).cover_img || ''
      form.value.category_id = Number((article as any).category_id || 0)
      // 从文章详情中提取标签（后端返回 Tag[] 对象数组）
      const articleTags = (article as any).tags || []
      form.value.tag_list = articleTags.map((t: any) => ({
        id: t.id ?? t.ID,
        name: t.name,
      }))
    }
  } catch (error) {
    ElMessage.error('获取文章失败')
    router.push('/')
  }
}

/** 获取所有分类 */
const fetchAllCategories = async () => {
  try {
    const list = await tagsApi.getCategories()
    categories.value = list
  } catch {
    // ignore
  }
}

/** 获取所有标签 */
const fetchAllTags = async () => {
  try {
    const tags = await tagsApi.getTags()
    allTags.value = tags
    // 推荐标签 = 所有标签中未被选中的前 10 个
    const selectedIds = new Set(form.value.tag_list.map(t => t.id))
    suggestedTags.value = tags
      .filter(t => !selectedIds.has(t.id))
      .slice(0, 10)
  } catch (error) {
    // ignore
  }
}

/** 发布/更新文章 */
const publishArticle = async () => {
  if (!formRef.value) return
  // 使用 Promise 形式的 validate，避免回调嵌套问题
  let valid = false
  try {
    valid = await formRef.value.validate()
  } catch {
    valid = false
  }
  if (!valid) return

  // ⚠️ 防御式前置校验：确保关键字段非空
  if (!form.value.title?.trim()) {
    ElMessage.warning('请输入文章标题')
    return
  }
  if (!form.value.content?.trim() || form.value.content.trim().length < 10) {
    ElMessage.warning('请输入至少 10 个字符的文章内容')
    return
  }

  saving.value = true
  try {
    form.value.preview = generatePreview(form.value.content)
    // 从 tag_list 中提取 numeric ID 数组，过滤掉无效值
    const tag_ids = form.value.tag_list
      .map((t: any) => {
        if (typeof t === 'number') return t
        return t ? t.id : null
      })
      .map((id: any) => Number(id))
      .filter((id: number) => Number.isInteger(id) && id > 0)

    const basePayload = {
      title: form.value.title.trim(),
      content: form.value.content,
      preview: form.value.preview,
      tag_ids,
      cover_img: form.value.cover_img,
    }

    if (isEditing.value) {
      await store.updateArticle(articleId.value, {
        ...basePayload,
        category_id: form.value.category_id ? Number(form.value.category_id) : 0,
      })
      ElMessage.success('文章更新成功')
      router.push(`/article/${articleId.value}`)
    } else {
      const created = await store.createArticle({
        ...basePayload,
        category_id: form.value.category_id ? Number(form.value.category_id) : null,
      })
      ElMessage.success('文章发布成功')
      localStorage.removeItem('article_draft')
      const createdId = getObjectId(created)
      router.push(createdId ? `/article/${createdId}` : '/')
    }
  } catch (error: any) {
    ElMessage.error(isEditing.value ? '更新失败' : '发布失败')
  } finally {
    saving.value = false
  }
}

/** 存为草稿（暂存到本地） */
const saveDraft = () => {
  const draft = {
    title: form.value.title,
    content: form.value.content,
    cover_img: form.value.cover_img,
    tag_list: form.value.tag_list,
    savedAt: new Date().toISOString(),
  }
  localStorage.setItem('article_draft', JSON.stringify(draft))
  ElMessage.success('草稿已保存到本地')
}

/** 加载本地草稿 */
const loadDraft = () => {
  const saved = localStorage.getItem('article_draft')
  if (saved) {
    try {
      const draft = JSON.parse(saved)
      ElMessageBox.confirm(
        `检测到本地草稿（保存于 ${new Date(draft.savedAt).toLocaleString('zh-CN')}），是否恢复？`,
        '恢复草稿',
        {
          confirmButtonText: '恢复',
          cancelButtonText: '放弃',
          type: 'info',
        }
      ).then(() => {
        form.value.title = draft.title || ''
        form.value.content = draft.content || ''
        form.value.cover_img = draft.cover_img || ''
        form.value.tag_list = draft.tag_list || []
        localStorage.removeItem('article_draft')
      }).catch(() => {
        localStorage.removeItem('article_draft')
      })
    } catch (e) {
      // ignore
    }
  }
}

onMounted(() => {
  if (isEditing.value) {
    fetchArticleForEdit()
  } else {
    loadDraft()
  }
  fetchAllTags()
  fetchAllCategories()
})
</script>

<style scoped lang="scss">
.main-container {
  max-width: 860px;
  margin: 0 auto;
  padding: 32px 24px 48px;
}

.editor-header {
  margin-bottom: 32px;

  .page-title {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 28px;
    font-weight: 700;
    color: #1a1a1a;
    margin: 0;
  }
}

.editor-form {
  :deep(.el-form-item__label) {
    font-weight: 600;
    font-size: 15px;
    color: #333;
  }
}

.title-input {
  :deep(.el-input__inner) {
    font-size: 22px;
    font-weight: 600;
    padding: 12px 0;
    border: none;
    border-bottom: 2px solid #f0f0f0;
    border-radius: 0;

    &:focus {
      border-bottom-color: #667eea;
    }
  }
}

.cover-upload {
  .cover-uploader {
    :deep(.el-upload) {
      cursor: pointer;
      border-radius: 12px;
      overflow: hidden;
      border: 2px dashed #ddd;
      transition: border-color 0.3s;

      &:hover {
        border-color: #667eea;
      }
    }
  }

  .cover-preview {
    position: relative;
    width: 320px;
    height: 180px;

    .preview-img {
      width: 100%;
      height: 100%;
    }

    .cover-overlay {
      position: absolute;
      inset: 0;
      background: rgba(0, 0, 0, 0.5);
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 8px;
      color: #fff;
      opacity: 0;
      transition: opacity 0.3s;
    }

    &:hover .cover-overlay {
      opacity: 1;
    }

    .remove-cover {
      position: absolute;
      top: 8px;
      right: 8px;
    }
  }

  .cover-placeholder {
    width: 320px;
    height: 180px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: #999;
    background: #fafafa;

    span {
      font-size: 14px;
    }

    .hint {
      font-size: 12px;
      color: #ccc;
    }
  }
}

.tags-section {
  width: 100%;

  .selected-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;
  }

  .tag-input-wrapper {
    margin-bottom: 12px;

    .tag-select {
      max-width: 400px;
    }
  }

  .suggested-tags {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;

    .suggested-label {
      font-size: 13px;
      color: #999;
    }

    .suggested-item {
      cursor: pointer;
      transition: all 0.2s;

      &:hover {
        transform: scale(1.05);
      }
    }
  }
}

.content-editor {
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  overflow: hidden;

  .editor-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    background: #fafafa;
    border-bottom: 1px solid #eee;

    .toolbar-right {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }

  .editor-body {
    min-height: 400px;
  }

  .editor-textarea {
    :deep(.el-textarea__inner) {
      border: none;
      border-radius: 0;
      padding: 16px 20px;
      font-size: 15px;
      line-height: 1.8;
      box-shadow: none;
      font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
    }
  }

  .editor-preview {
    padding: 16px 20px;
    min-height: 400px;
    font-size: 15px;
    line-height: 1.8;
    color: #333;

    :deep(h1), :deep(h2), :deep(h3) {
      margin: 20px 0 10px;
      color: #1a1a1a;
    }

    :deep(strong) {
      font-weight: 600;
    }

    :deep(code) {
      background: #f5f5f5;
      padding: 2px 6px;
      border-radius: 4px;
      font-size: 0.9em;
    }

    :deep(pre) {
      background: #282c34;
      color: #abb2bf;
      padding: 16px;
      border-radius: 8px;
      overflow-x: auto;
    }

    :deep(blockquote) {
      border-left: 4px solid #667eea;
      margin: 12px 0;
      padding: 8px 16px;
      background: #f8f9ff;
      border-radius: 0 6px 6px 0;
    }

    :deep(a) {
      color: #667eea;
    }
  }
}

.submit-actions {
  .actions-wrapper {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
    padding-top: 16px;
    border-top: 1px solid #f0f0f0;
    width: 100%;
  }
}

@media (max-width: 768px) {
  .main-container {
    padding: 16px 16px 32px;
  }

  .editor-header .page-title {
    font-size: 22px;
  }

  .cover-upload {
    .cover-preview,
    .cover-placeholder {
      width: 100%;
      height: 160px;
    }
  }

  .submit-actions .actions-wrapper {
    flex-direction: column;

    .el-button {
      width: 100%;
    }
  }
}
</style>
