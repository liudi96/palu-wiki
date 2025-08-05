'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import UserLayout from '@/components/UserLayout';
import { articleAPI, type Article } from '@/lib/api';
import Link from 'next/link';
import {
  DocumentTextIcon,
  SparklesIcon,
  EyeIcon,
  CalendarIcon,
  UserIcon,
  FolderIcon,
  ArrowLeftIcon,
  ShareIcon,
  BookmarkIcon,
} from '@heroicons/react/24/outline';

export default function ArticleDetailPage() {
  const params = useParams();
  const articleId = params.id as string;
  const [article, setArticle] = useState<Article | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchArticle = async () => {
      if (!articleId) return;
      
      setLoading(true);
      try {
        const response = await articleAPI.getArticle(parseInt(articleId));
        setArticle(response.data);
      } catch (error: any) {
        console.error('Failed to fetch article:', error);
        if (error.response?.status === 404) {
          setError('文章不存在');
        } else {
          setError('加载文章失败');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchArticle();
  }, [articleId]);

  const handleShare = () => {
    if (navigator.share) {
      navigator.share({
        title: article?.title,
        url: window.location.href,
      });
    } else {
      // 复制链接到剪贴板
      navigator.clipboard.writeText(window.location.href);
      alert('链接已复制到剪贴板');
    }
  };

  const handleBookmark = () => {
    // 这里可以实现收藏功能
    alert('收藏功能待实现');
  };

  if (loading) {
    return (
      <UserLayout>
        <div className="max-w-4xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
          <div className="flex items-center justify-center h-64">
            <div className="text-lg text-gray-500">加载中...</div>
          </div>
        </div>
      </UserLayout>
    );
  }

  if (error || !article) {
    return (
      <UserLayout>
        <div className="max-w-4xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
          <div className="text-center py-12">
            <DocumentTextIcon className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">{error || '文章不存在'}</h3>
            <p className="mt-1 text-sm text-gray-500">
              请检查链接是否正确，或返回首页浏览其他内容
            </p>
            <div className="mt-6">
              <Link
                href="/articles"
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700"
              >
                <ArrowLeftIcon className="h-4 w-4 mr-2" />
                返回攻略列表
              </Link>
            </div>
          </div>
        </div>
      </UserLayout>
    );
  }

  return (
    <UserLayout>
      <div className="max-w-4xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
        {/* 返回链接 */}
        <div className="mb-6">
          <Link
            href="/articles"
            className="inline-flex items-center text-sm text-gray-500 hover:text-gray-700"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            返回攻略列表
          </Link>
        </div>

        {/* 文章内容 */}
        <article className="bg-white shadow-lg rounded-lg overflow-hidden">
          {/* 文章头部 */}
          <div className="px-6 py-8 border-b border-gray-200">
            {/* 分类和标签 */}
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2 text-sm text-gray-500">
                  <FolderIcon className="h-4 w-4" />
                  <Link
                    href={`/categories/${article.category?.id}`}
                    className="hover:text-primary-600"
                  >
                    {article.category?.name || '未分类'}
                  </Link>
                </div>
                {article.is_ai_generated && (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                    <SparklesIcon className="h-3 w-3 mr-1" />
                    AI生成
                  </span>
                )}
              </div>
              
              {/* 操作按钮 */}
              <div className="flex items-center space-x-2">
                <button
                  onClick={handleShare}
                  className="p-2 text-gray-400 hover:text-gray-600 rounded-full hover:bg-gray-100"
                  title="分享"
                >
                  <ShareIcon className="h-5 w-5" />
                </button>
                <button
                  onClick={handleBookmark}
                  className="p-2 text-gray-400 hover:text-gray-600 rounded-full hover:bg-gray-100"
                  title="收藏"
                >
                  <BookmarkIcon className="h-5 w-5" />
                </button>
              </div>
            </div>

            {/* 标题 */}
            <h1 className="text-3xl font-bold text-gray-900 mb-4">
              {article.title}
            </h1>

            {/* 摘要 */}
            {article.summary && (
              <p className="text-lg text-gray-600 mb-6 leading-relaxed">
                {article.summary}
              </p>
            )}

            {/* 元信息 */}
            <div className="flex items-center space-x-6 text-sm text-gray-500">
              <div className="flex items-center">
                <UserIcon className="h-4 w-4 mr-1" />
                <span>{article.author?.nickname || article.author?.username || '匿名'}</span>
              </div>
              <div className="flex items-center">
                <CalendarIcon className="h-4 w-4 mr-1" />
                <span>{new Date(article.created_at).toLocaleString()}</span>
              </div>
              <div className="flex items-center">
                <EyeIcon className="h-4 w-4 mr-1" />
                <span>{article.view_count || 0} 次浏览</span>
              </div>
              {article.updated_at !== article.created_at && (
                <div className="text-xs text-gray-400">
                  更新于 {new Date(article.updated_at).toLocaleString()}
                </div>
              )}
            </div>
          </div>

          {/* 文章正文 */}
          <div className="px-6 py-8">
            <div 
              className="prose prose-lg max-w-none prose-headings:text-gray-900 prose-p:text-gray-700 prose-a:text-primary-600 prose-strong:text-gray-900"
              dangerouslySetInnerHTML={{
                __html: article.content.replace(/\n/g, '<br />').replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
              }}
            />
          </div>
        </article>

        {/* 相关推荐 */}
        <div className="mt-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">
            📖 相关攻略推荐
          </h2>
          <div className="bg-white rounded-lg shadow p-6">
            <p className="text-gray-500 text-center py-8">
              相关推荐功能开发中...
            </p>
          </div>
        </div>

        {/* 评论区域 */}
        <div className="mt-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">
            💬 评论讨论
          </h2>
          <div className="bg-white rounded-lg shadow p-6">
            <p className="text-gray-500 text-center py-8">
              评论功能开发中，敬请期待...
            </p>
          </div>
        </div>

        {/* 返回顶部按钮 */}
        <div className="mt-8 text-center">
          <button
            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
            className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
          >
            返回顶部
          </button>
        </div>
      </div>
    </UserLayout>
  );
}