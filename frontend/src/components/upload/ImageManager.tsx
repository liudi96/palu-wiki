'use client'

import { useState, useEffect } from 'react'
import { toast } from 'react-hot-toast'
import ImageUpload from './ImageUpload'
import { uploadAPI } from '@/lib/api'

interface FileItem {
  id: number
  file_name: string
  stored_name: string
  file_size: number
  file_type: string
  file_url: string
  width: number
  height: number
  uploader: {
    id: number
    username: string
    nickname: string
  }
  description: string
  created_at: string
}

interface ImageManagerProps {
  onSelectImage?: (imageUrl: string) => void
  showUpload?: boolean
  showSelect?: boolean
}

export default function ImageManager({ 
  onSelectImage, 
  showUpload = true, 
  showSelect = true 
}: ImageManagerProps) {
  const [files, setFiles] = useState<FileItem[]>([])
  const [loading, setLoading] = useState(false)
  const [selectedImage, setSelectedImage] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  const fetchFiles = async (pageNum = 1) => {
    setLoading(true)
    try {
      const result = await uploadAPI.getFiles({
        type: 'image',
        page: pageNum,
        page_size: 12
      })

      if (result.success) {
        setFiles(result.data || [])
        setTotalPages(result.pagination.total_pages)
        setPage(pageNum)
      } else {
        throw new Error(result.error || '获取文件列表失败')
      }
    } catch (error) {
      console.error('Fetch files error:', error)
      toast.error(error instanceof Error ? error.message : '获取文件列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchFiles()
  }, [])

  const handleUploadSuccess = (imageUrl: string) => {
    // 刷新文件列表
    fetchFiles(page)
    if (onSelectImage) {
      onSelectImage(imageUrl)
    }
  }

  const handleSelectImage = (file: FileItem) => {
    const imageUrl = `http://localhost:8080${file.file_url}`
    setSelectedImage(imageUrl)
    if (onSelectImage) {
      onSelectImage(imageUrl)
    }
  }

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    })
  }

  return (
    <div className="space-y-6">
      {showUpload && (
        <div>
          <h3 className="text-lg font-medium mb-3">上传新图片</h3>
          <ImageUpload onUploadSuccess={handleUploadSuccess} />
        </div>
      )}

      <div>
        <div className="flex justify-between items-center mb-3">
          <h3 className="text-lg font-medium">图片库</h3>
          <button
            onClick={() => fetchFiles(page)}
            disabled={loading}
            className="px-3 py-1 text-sm bg-gray-500 text-white rounded hover:bg-gray-600 disabled:opacity-50"
          >
            {loading ? '刷新中...' : '刷新'}
          </button>
        </div>

        {loading ? (
          <div className="flex justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          </div>
        ) : files.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            还没有上传任何图片
          </div>
        ) : (
          <>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
              {files.map((file) => (
                <div
                  key={file.id}
                  className={`
                    border rounded-lg overflow-hidden cursor-pointer transition-all
                    ${selectedImage === `http://localhost:8080${file.file_url}` 
                      ? 'border-blue-500 ring-2 ring-blue-200' 
                      : 'border-gray-200 hover:border-gray-300'
                    }
                  `}
                  onClick={() => showSelect && handleSelectImage(file)}
                >
                  <div className="aspect-square relative">
                    <img
                      src={`http://localhost:8080${file.file_url}`}
                      alt={file.file_name}
                      className="w-full h-full object-cover"
                    />
                    {showSelect && (
                      <div className="absolute inset-0 bg-black bg-opacity-0 hover:bg-opacity-20 flex items-center justify-center transition-all">
                        <span className="text-white opacity-0 hover:opacity-100 transition-opacity">
                          点击选择
                        </span>
                      </div>
                    )}
                  </div>
                  
                  <div className="p-3">
                    <p className="text-sm font-medium truncate" title={file.file_name}>
                      {file.file_name}
                    </p>
                    <p className="text-xs text-gray-500">
                      {file.width} × {file.height}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatFileSize(file.file_size)}
                    </p>
                    <p className="text-xs text-gray-400">
                      {formatDate(file.created_at)}
                    </p>
                    {showSelect && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          const markdown = `![${file.file_name}](http://localhost:8080${file.file_url})`
                          navigator.clipboard.writeText(markdown)
                          toast.success('Markdown代码已复制')
                        }}
                        className="mt-2 w-full px-2 py-1 text-xs bg-green-500 text-white rounded hover:bg-green-600"
                      >
                        复制Markdown
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>

            {totalPages > 1 && (
              <div className="flex justify-center items-center gap-2 mt-6">
                <button
                  onClick={() => fetchFiles(page - 1)}
                  disabled={page <= 1 || loading}
                  className="px-3 py-2 text-sm bg-gray-500 text-white rounded hover:bg-gray-600 disabled:opacity-50"
                >
                  上一页
                </button>
                <span className="text-sm text-gray-600">
                  第 {page} 页，共 {totalPages} 页
                </span>
                <button
                  onClick={() => fetchFiles(page + 1)}
                  disabled={page >= totalPages || loading}
                  className="px-3 py-2 text-sm bg-gray-500 text-white rounded hover:bg-gray-600 disabled:opacity-50"
                >
                  下一页
                </button>
              </div>
            )}
          </>
        )}
      </div>

      {selectedImage && showSelect && (
        <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <p className="text-sm text-blue-800">
            已选择图片: <span className="font-mono text-xs">{selectedImage}</span>
          </p>
        </div>
      )}
    </div>
  )
}