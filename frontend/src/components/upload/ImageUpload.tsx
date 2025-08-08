'use client'

import { useState, useRef } from 'react'
import { toast } from 'react-hot-toast'
import { uploadAPI } from '@/lib/api'

interface ImageUploadProps {
  onUploadSuccess?: (imageUrl: string) => void
  onUploadError?: (error: string) => void
  accept?: string
  maxSize?: number // MB
  className?: string
}

interface UploadedFile {
  id: number
  file_name: string
  file_url: string
  width: number
  height: number
  file_size: number
}

export default function ImageUpload({
  onUploadSuccess,
  onUploadError,
  accept = "image/*",
  maxSize = 10,
  className = ""
}: ImageUploadProps) {
  const [uploading, setUploading] = useState(false)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [uploadedFile, setUploadedFile] = useState<UploadedFile | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    // 检查文件大小
    if (file.size > maxSize * 1024 * 1024) {
      const error = `文件大小不能超过 ${maxSize}MB`
      toast.error(error)
      onUploadError?.(error)
      return
    }

    // 检查文件类型
    if (!file.type.startsWith('image/')) {
      const error = '请选择图片文件'
      toast.error(error)
      onUploadError?.(error)
      return
    }

    // 显示预览
    const reader = new FileReader()
    reader.onload = (e) => {
      setPreviewUrl(e.target?.result as string)
    }
    reader.readAsDataURL(file)

    // 上传文件
    uploadFile(file)
  }

  const uploadFile = async (file: File) => {
    setUploading(true)

    try {
      const description = `上传的图片: ${file.name}`
      const result = await uploadAPI.uploadImage(file, description)

      if (result.success) {
        const uploadedFile: UploadedFile = result.data
        setUploadedFile(uploadedFile)
        
        // 构建完整的图片URL
        const imageUrl = `http://localhost:8080${uploadedFile.file_url}`
        
        toast.success('图片上传成功！')
        onUploadSuccess?.(imageUrl)
      } else {
        throw new Error(result.error || '上传失败')
      }
    } catch (error) {
      console.error('Upload error:', error)
      const errorMessage = error instanceof Error ? error.message : '上传失败'
      toast.error(errorMessage)
      onUploadError?.(errorMessage)
      setPreviewUrl(null)
    } finally {
      setUploading(false)
    }
  }

  const handleClick = () => {
    fileInputRef.current?.click()
  }

  const handleRemove = () => {
    setPreviewUrl(null)
    setUploadedFile(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  return (
    <div className={`image-upload ${className}`}>
      <input
        ref={fileInputRef}
        type="file"
        accept={accept}
        onChange={handleFileSelect}
        className="hidden"
      />

      {previewUrl ? (
        <div className="relative">
          <div className="border-2 border-dashed border-gray-300 rounded-lg p-4 bg-gray-50">
            <img
              src={previewUrl}
              alt="预览图片"
              className="max-w-full max-h-48 mx-auto rounded"
            />
            {uploadedFile && (
              <div className="mt-2 text-sm text-gray-600">
                <p>文件名: {uploadedFile.file_name}</p>
                <p>尺寸: {uploadedFile.width} × {uploadedFile.height}</p>
                <p>大小: {(uploadedFile.file_size / 1024).toFixed(2)} KB</p>
              </div>
            )}
          </div>
          
          <div className="flex gap-2 mt-2">
            <button
              type="button"
              onClick={handleClick}
              disabled={uploading}
              className="px-3 py-1 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
            >
              {uploading ? '上传中...' : '重新选择'}
            </button>
            <button
              type="button"
              onClick={handleRemove}
              disabled={uploading}
              className="px-3 py-1 text-sm bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50"
            >
              删除
            </button>
            {uploadedFile && (
              <button
                type="button"
                onClick={() => {
                  const markdown = `![${uploadedFile.file_name}](http://localhost:8080${uploadedFile.file_url})`
                  navigator.clipboard.writeText(markdown)
                  toast.success('Markdown代码已复制到剪贴板')
                }}
                className="px-3 py-1 text-sm bg-green-500 text-white rounded hover:bg-green-600"
              >
                复制Markdown
              </button>
            )}
          </div>
        </div>
      ) : (
        <div
          onClick={handleClick}
          className={`
            border-2 border-dashed border-gray-300 rounded-lg p-8 text-center cursor-pointer
            hover:border-blue-400 hover:bg-blue-50 transition-colors
            ${uploading ? 'pointer-events-none opacity-50' : ''}
          `}
        >
          {uploading ? (
            <div className="flex flex-col items-center gap-2">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
              <p className="text-gray-600">上传中...</p>
            </div>
          ) : (
            <div className="flex flex-col items-center gap-2">
              <svg className="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              <p className="text-gray-600">
                点击上传图片或拖拽到此处
              </p>
              <p className="text-sm text-gray-500">
                支持 JPG, PNG, GIF 格式，最大 {maxSize}MB
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}