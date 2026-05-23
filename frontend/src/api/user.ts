import http from './http'
import type { ProfileResponse } from '@/types/common'

/** 关注/粉丝列表响应 */
export interface FollowListResponse {
    data: Array<{
        id: number
        username: string
        image: string
        bio: string
    }>
    total: number
    page: number
    page_size: number
    total_pages: number
}

/** 关注数/粉丝数响应 */
export interface FollowCountsResponse {
    following_count: number
    followers_count: number
}

function normalizeUser(raw: any) {
    if (!raw) return raw
    if (raw.ID !== undefined && raw.id === undefined) raw.id = raw.ID
    if (raw.CreatedAt !== undefined && raw.created_at === undefined) raw.created_at = raw.CreatedAt
    if (raw.UpdatedAt !== undefined && raw.updated_at === undefined) raw.updated_at = raw.UpdatedAt
    return raw
}

export const userApi = {
    /** 获取用户资料（按用户名） */
    async getProfile(username: string): Promise<ProfileResponse> {
        const res = await http.get<ProfileResponse>(`/api/v1/profiles/${username}`)
        return res.data
    },

    /** 关注用户（按用户名 — 保留兼容） */
    async followUser(username: string): Promise<ProfileResponse> {
        const res = await http.post<ProfileResponse>(`/api/v1/profiles/${username}/follow`)
        return res.data
    },

    /** 取消关注用户（按用户名 — 保留兼容） */
    async unfollowUser(username: string): Promise<ProfileResponse> {
        const res = await http.delete<ProfileResponse>(`/api/v1/profiles/${username}/follow`)
        return res.data
    },

    /** 获取用户资料（按用户 ID） */
    async getUserById(userId: number): Promise<ProfileResponse> {
        const res = await http.get<ProfileResponse>(`/api/v1/user/${userId}`)
        normalizeUser((res.data as any).user)
        return res.data
    },

    /** 关注用户（按用户 ID — 匹配后端 /user/:id/follow 路由） */
    async followUserById(userId: number): Promise<any> {
        const res = await http.post(`/api/v1/user/${userId}/follow`)
        return res.data
    },

    /** 取消关注用户（按用户 ID — 匹配后端 /user/:id/follow 路由） */
    async unfollowUserById(userId: number): Promise<any> {
        // 后端关注路由只注册了 POST（toggle 语义），没有 DELETE
        // 取消关注同样调用 POST /user/:id/follow，后端自动 toggle
        const res = await http.post(`/api/v1/user/${userId}/follow`)
        return res.data
    },

    /** 获取关注状态 */
    async getFollowStatus(userId: number): Promise<{ is_following: boolean }> {
        const res = await http.get<{ is_following: boolean }>(`/api/v1/user/${userId}/follow`)
        return res.data
    },

    /** 获取用户的关注数和粉丝数 */
    async getFollowCounts(userId: number): Promise<FollowCountsResponse> {
        const res = await http.get<FollowCountsResponse>(`/api/v1/user/${userId}/follow/counts`)
        return res.data
    },

    /** 获取用户的关注列表 */
    async getFollowing(userId: number, params?: { page?: number; page_size?: number }): Promise<FollowListResponse> {
        const res = await http.get<FollowListResponse>(`/api/v1/user/${userId}/following`, { params })
        res.data.data = (res.data.data || []).map(normalizeUser)
        return res.data
    },

    /** 获取用户的粉丝列表 */
    async getFollowers(userId: number, params?: { page?: number; page_size?: number }): Promise<FollowListResponse> {
        const res = await http.get<FollowListResponse>(`/api/v1/user/${userId}/followers`, { params })
        res.data.data = (res.data.data || []).map(normalizeUser)
        return res.data
    },
}
