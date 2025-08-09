import axios from 'axios';
import Cookies from 'js-cookie';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// 创建axios实例
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 120000, // 放宽到120s，避免AI生成超时
});

// 请求拦截器 - 自动添加token
api.interceptors.request.use(
  (config) => {
    const token = Cookies.get('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理错误
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token过期，清除cookie并跳转到登录页
      Cookies.remove('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export interface LoginData {
  username: string;
  password: string;
}

export interface User {
  id: number;
  username: string;
  email: string;
  nickname: string;
  role: string;
  status: string;
  created_at: string;
}

export interface Article {
  id: number;
  title: string;
  content: string;
  summary: string;
  status: string;
  is_ai_generated: boolean;
  author: User;
  category: Category;
  view_count: number;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: number;
  name: string;
  description: string;
  sort_order: number;
  article_count: number;
}

export interface PaginationResponse<T> {
  data: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

export interface DashboardStats {
  total_users: number;
  total_articles: number;
  total_categories: number;
  total_comments: number;
  ai_articles: number;
  pending_articles: number;
  published_articles: number;
  draft_articles: number;
}

// 认证相关API
export const authAPI = {
  login: async (data: LoginData) => {
    const response = await api.post('/api/v1/auth/login', data);
    return response.data;
  },
  
  getProfile: async () => {
    const response = await api.get('/api/v1/user/profile');
    return response.data;
  },
};

// 管理后台API
export const adminAPI = {
  // 仪表板
  getDashboard: async () => {
    const response = await api.get('/api/v1/admin/dashboard');
    return response.data;
  },
  
  getSystemInfo: async () => {
    const response = await api.get('/api/v1/admin/system');
    return response.data;
  },
  
  // 用户管理
  getUsers: async (params?: { page?: number; page_size?: number; q?: string; role?: string; status?: string }) => {
    const response = await api.get('/api/v1/admin/users', { params });
    return response.data;
  },
  
  updateUserStatus: async (userId: number, data: { status: string; role?: string }) => {
    const response = await api.put(`/api/v1/admin/users/${userId}/status`, data);
    return response.data;
  },
  
  // 文章管理
  getArticles: async (params?: { 
    page?: number; 
    page_size?: number; 
    status?: string; 
    category_id?: string; 
    author_id?: string; 
    is_ai_generated?: string;
  }) => {
    const response = await api.get('/api/v1/admin/articles', { params });
    return response.data;
  },
  
  updateArticleStatus: async (articleId: number, data: { status: string; reason?: string }) => {
    const response = await api.put(`/api/v1/admin/articles/${articleId}/status`, data);
    return response.data;
  },
};

// 文章相关API
export const articleAPI = {
  search: async (params: {
    q: string;
    page?: number;
    page_size?: number;
    category_id?: string;
    author_id?: string;
    is_ai_generated?: string;
    status?: string;
    sort?: string;
    order?: string;
  }) => {
    const response = await api.get('/api/v1/articles/search', { params });
    return response.data;
  },
  
  getArticles: async (params?: { page?: number; limit?: number }) => {
    const response = await api.get('/api/v1/articles', { params });
    return response.data;
  },
  
  getArticle: async (id: number) => {
    const response = await api.get(`/api/v1/articles/${id}`);
    return response.data;
  },

  createArticle: async (data: {
    title: string;
    content: string;
    summary: string;
    category_id: number;
    status: string;
  }) => {
    const response = await api.post('/api/v1/articles', data);
    return response.data;
  },
};

// AI相关API
export const aiAPI = {
  generateContent: async (data: {
    title: string;
    category_id?: number;
  }) => {
    const response = await api.post('/api/v1/ai/generate', data);
    return response.data;
  },
};

// 分类相关API
export const categoryAPI = {
  getCategories: async () => {
    const response = await api.get('/api/v1/categories');
    return response.data;
  },
};

// 文件上传相关API
export const uploadAPI = {
  uploadImage: async (file: File, description?: string) => {
    const formData = new FormData();
    formData.append('image', file);
    if (description) {
      formData.append('description', description);
    }
    
    const response = await api.post('/api/v1/upload/image', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  uploadFile: async (file: File, description?: string) => {
    const formData = new FormData();
    formData.append('file', file);
    if (description) {
      formData.append('description', description);
    }
    
    const response = await api.post('/api/v1/upload/file', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  getFiles: async (params?: {
    page?: number;
    page_size?: number;
    type?: string;
    uploader_id?: string;
  }) => {
    const response = await api.get('/api/v1/upload/files', { params });
    return response.data;
  },

  getFile: async (id: number) => {
    const response = await api.get(`/api/v1/upload/files/${id}`);
    return response.data;
  },

  updateFile: async (id: number, description: string) => {
    const formData = new FormData();
    formData.append('description', description);
    
    const response = await api.put(`/api/v1/upload/files/${id}`, formData);
    return response.data;
  },

  deleteFile: async (id: number) => {
    const response = await api.delete(`/api/v1/upload/files/${id}`);
    return response.data;
  },
};

// 导出api实例作为apiClient
export const apiClient = api;
export default api;