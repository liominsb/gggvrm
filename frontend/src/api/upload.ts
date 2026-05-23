import http from './http'
import type { UploadResponse } from '@/types/common'

/** 后端基地址，用于拼接图片完整路径 */
const BASE_URL = import.meta.env.DEV ? '' : 'http://localhost:8080'

/**
 * 将图片相对路径转换为完整 URL
 * 后端返回如 ["/uploads/images/xxx.jpg"]，需要拼接域名
 */
export function getFullImageUrl(relativePath: string): string {
    if (!relativePath) return ''
    if (relativePath.startsWith('http')) return relativePath
    return `${BASE_URL}${relativePath}`
}

export const uploadApi = {
    /**
     * 批量上传图片
     * @param files File 数组
     * @returns 图片相对路径数组
     */
    async uploadImages(files: File[]): Promise<string[]> {
        const formData = new FormData()
        files.forEach((file) => {
            formData.append('files', file)
        })

        const res = await http.post<UploadResponse>('/api/v1/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
        })

        // 后端返回的是相对路径数组，转为完整 URL
        return res.data.urls.map((url) => getFullImageUrl(url))
    },
}