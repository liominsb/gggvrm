import http from './http'
import type { Tag, Category } from '@/types/common'

/** 将后端 gorm.Model 的大写 ID 转为小写 id */
function normalizeTag(raw: any): Tag {
    return {
        id: raw.id ?? raw.ID ?? 0,
        name: raw.name ?? '',
    }
}

function normalizeTags(raw: any[]): Tag[] {
    return (raw || []).map(normalizeTag)
}

function normalizeCategory(raw: any): Category {
    return {
        id: raw.id ?? raw.ID ?? 0,
        name: raw.name ?? '',
        slug: raw.slug ?? '',
    }
}

export const tagsApi = {
    /** 获取所有标签（后端直接返回 Tag[] 数组） */
    async getTags(): Promise<Tag[]> {
        const res = await http.get<any[]>('/api/v1/tags')
        return normalizeTags(res.data)
    },

    /** 创建标签 */
    async createTag(name: string): Promise<Tag> {
        const res = await http.post<any>('/api/v1/tag', { name })
        return normalizeTag(res.data)
    },

    /** 删除标签 */
    async deleteTag(id: number): Promise<void> {
        await http.delete(`/api/v1/tag/${id}`)
    },

    /** 获取所有分类 —— 后端直接返回 Category[] 数组 */
    async getCategories(): Promise<Category[]> {
        const res = await http.get<any>('/api/v1/categories')
        const data = res.data
        // 后端可能直接返回数组或包装在 { categories: [...] } 中
        const raw = Array.isArray(data) ? data : (data?.categories || [])
        return raw.map(normalizeCategory)
    },
}
