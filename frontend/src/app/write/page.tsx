'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Cookies from 'js-cookie';
import UserLayout from '@/components/UserLayout';
import { categoryAPI, articleAPI, aiAPI } from '@/lib/api';
import { toast } from 'react-hot-toast';

interface Category {
  id: number;
  name: string;
  description: string;
}

export default function WritePage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [categories, setCategories] = useState<Category[]>([]);
  const [formData, setFormData] = useState({
    title: '',
    content: '',
    summary: '',
    category_id: '',
    status: 'draft' as 'draft' | 'published'
  });
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    // 检查登录状态 - 使用Cookie而不是localStorage
    const token = Cookies.get('token');
    if (!token) {
      router.push('/login?redirect=/write');
      return;
    }
    setIsLoggedIn(true);
    
    // 加载分类数据
    loadCategories();
  }, [router]);

  const loadCategories = async () => {
    try {
      const response = await categoryAPI.getCategories();
      setCategories(response.data || []);
    } catch (error) {
      console.error('加载分类失败:', error);
      toast.error('加载分类失败');
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.title.trim()) {
      toast.error('请输入文章标题');
      return;
    }
    
    if (!formData.content.trim()) {
      toast.error('请输入文章内容');
      return;
    }
    
    if (!formData.category_id) {
      toast.error('请选择文章分类');
      return;
    }

    setIsLoading(true);
    
    try {
      const articleData = {
        ...formData,
        category_id: parseInt(formData.category_id)
      };
      
      const response = await articleAPI.createArticle(articleData);
      
      if (response.success) {
        toast.success(`文章${formData.status === 'published' ? '发布' : '保存'}成功！`);
        router.push(`/articles/${response.data.id}`);
      } else {
        throw new Error(response.error || '提交失败');
      }
    } catch (error: any) {
      console.error('提交文章失败:', error);
      toast.error(error?.response?.data?.error || '提交文章失败，请重试');
    } finally {
      setIsLoading(false);
    }
  };

  const handleAIGenerate = async () => {
    if (!formData.title.trim()) {
      toast.error('请先输入文章标题');
      return;
    }

    setIsLoading(true);
    try {
      const response = await aiAPI.generateContent({
        title: formData.title,
        category_id: formData.category_id ? parseInt(formData.category_id) : undefined
      });

      if (response.success) {
        const aiContent = response.data;
        setFormData(prev => ({
          ...prev,
          content: aiContent.content,
          summary: aiContent.summary
        }));
        toast.success('AI内容生成成功！');
      } else {
        throw new Error(response.error || 'AI生成失败');
      }
    } catch (error: any) {
      console.error('AI生成失败:', error);
      toast.error(error?.response?.data?.error || 'AI生成失败，请重试');
    } finally {
      setIsLoading(false);
    }
  };

  if (!isLoggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <UserLayout>
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="mb-6">
            <h1 className="text-3xl font-bold text-gray-900 mb-2">✍️ 发布攻略</h1>
            <p className="text-gray-600">分享你的游戏经验，帮助更多玩家成长</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* 标题输入 */}
            <div>
              <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-2">
                文章标题 *
              </label>
              <input
                type="text"
                id="title"
                name="title"
                value={formData.title}
                onChange={handleInputChange}
                placeholder="请输入吸引人的标题..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
              />
            </div>

            {/* 分类选择 */}
            <div>
              <label htmlFor="category_id" className="block text-sm font-medium text-gray-700 mb-2">
                文章分类 *
              </label>
              <select
                id="category_id"
                name="category_id"
                value={formData.category_id}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
              >
                <option value="">请选择分类</option>
                {categories.map((category) => (
                  <option key={category.id} value={category.id}>
                    {category.name}
                  </option>
                ))}
              </select>
            </div>

            {/* 内容编辑器 */}
            <div>
              <div className="flex justify-between items-center mb-2">
                <label htmlFor="content" className="block text-sm font-medium text-gray-700">
                  文章内容 *
                </label>
                <button
                  type="button"
                  onClick={handleAIGenerate}
                  disabled={isLoading}
                  className="px-4 py-2 bg-gradient-to-r from-purple-600 to-blue-600 text-white text-sm rounded-md hover:from-purple-700 hover:to-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                  <span>🤖</span>
                  {isLoading ? '生成中...' : 'AI智能生成'}
                </button>
              </div>
              <textarea
                id="content"
                name="content"
                value={formData.content}
                onChange={handleInputChange}
                placeholder="请输入文章内容，支持Markdown格式..."
                rows={15}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                required
              />
              <p className="text-xs text-gray-500 mt-1">
                支持 Markdown 格式：**粗体**、*斜体*、### 标题、- 列表等
              </p>
            </div>

            {/* 摘要输入 */}
            <div>
              <label htmlFor="summary" className="block text-sm font-medium text-gray-700 mb-2">
                文章摘要
              </label>
              <textarea
                id="summary"
                name="summary"
                value={formData.summary}
                onChange={handleInputChange}
                placeholder="简要描述文章内容，帮助读者快速了解..."
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>

            {/* 发布选项 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                发布状态
              </label>
              <div className="space-y-2">
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="status"
                    value="draft"
                    checked={formData.status === 'draft'}
                    onChange={handleInputChange}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">保存为草稿</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="status"
                    value="published"
                    checked={formData.status === 'published'}
                    onChange={handleInputChange}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">立即发布</span>
                </label>
              </div>
            </div>

            {/* 提交按钮 */}
            <div className="flex gap-4 pt-4">
              <button
                type="submit"
                disabled={isLoading}
                className="flex-1 bg-blue-600 text-white py-3 px-6 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
              >
                {isLoading ? (
                  <span className="flex items-center justify-center gap-2">
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                    提交中...
                  </span>
                ) : (
                  formData.status === 'published' ? '发布文章' : '保存草稿'
                )}
              </button>
              
              <button
                type="button"
                onClick={() => router.back()}
                className="px-6 py-3 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
              >
                取消
              </button>
            </div>
          </form>
        </div>

        {/* 写作提示 */}
        <div className="mt-8 bg-blue-50 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-blue-900 mb-3">📝 写作小贴士</h3>
          <ul className="space-y-2 text-blue-800">
            <li>• 使用清晰的标题和副标题来组织内容</li>
            <li>• 添加截图和图片可以让攻略更直观</li>
            <li>• 分步骤说明操作流程，便于读者跟随</li>
            <li>• 可以使用AI功能快速生成初稿，再进行个性化编辑</li>
            <li>• 保存草稿可以随时回来继续编辑</li>
          </ul>
        </div>
      </div>
    </UserLayout>
  );
}