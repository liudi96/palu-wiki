#!/usr/bin/env python3
"""
简化的gRPC测试服务器 - 用于测试Go客户端集成
不依赖复杂的AI服务，直接返回模拟数据
"""

import asyncio
import time
from typing import Iterator

import grpc
from grpc import aio
from loguru import logger

# 导入生成的gRPC代码
import sys
import os
sys.path.append(os.path.join(os.path.dirname(__file__), 'ai-service'))

try:
    from ai_service.app.grpc_generated import ai_service_pb2, ai_service_pb2_grpc
except ImportError:
    # 如果无法导入，创建简单的模拟实现
    logger.warning("无法导入gRPC生成代码，将创建模拟服务")
    
    # 这里需要使用原始的proto文件生成代码
    # 简单起见，我们先创建一个基本的测试服务


class MockAIServiceGRPC:
    """模拟AI服务gRPC实现"""
    
    def __init__(self):
        self.request_count = 0
        self.start_time = time.time()
    
    async def GenerateContent(self, request, context):
        """生成内容"""
        self.request_count += 1
        logger.info(f"🤖 收到内容生成请求: {request.prompt[:50]}...")
        
        # 模拟处理时间
        await asyncio.sleep(0.5)
        
        from ai_service.app.grpc_generated.ai_service_pb2 import ContentResponse
        
        return ContentResponse(
            content=f"这是一个关于「{request.prompt}」的模拟回答。帕鲁游戏是一款非常有趣的生存建造游戏，玩家可以捕捉和培养各种可爱的生物帕鲁。游戏具有丰富的内容和出色的画面表现。",
            summary="帕鲁游戏简介和特色介绍",
            tags=["帕鲁", "游戏", "攻略", "AI生成"],
            model_used="mock-ai-model",
            tokens_used=150,
            generation_time=0.5,
            confidence=0.95
        )
    
    async def GenerateArticle(self, request, context):
        """生成文章"""
        self.request_count += 1
        logger.info(f"📝 收到文章生成请求: {request.title}")
        
        # 模拟处理时间
        await asyncio.sleep(1.0)
        
        from ai_service.app.grpc_generated.ai_service_pb2 import ArticleResponse
        
        content = f"""# {request.title}

## 简介

{request.title}是一个关于帕鲁游戏的详细指南。帕鲁是一款结合了生存、建造和冒险元素的游戏。

## 主要特点

1. **丰富的帕鲁种类** - 游戏中有超过100种不同的帕鲁可以捕获和培养
2. **多样的游戏玩法** - 包括建造、战斗、探索等多种游戏模式
3. **精美的画面** - 采用卡通风格的3D画面，色彩丰富

## 入门建议

- 新手玩家建议先熟悉基础操作
- 合理分配资源，优先建造必需的设施
- 多探索不同区域，发现稀有的帕鲁

## 结语

{request.title}希望能帮助玩家更好地享受帕鲁世界的乐趣。

*本文由AI生成，仅供参考*
"""
        
        return ArticleResponse(
            title=request.title,
            content=content,
            summary="详细介绍了帕鲁游戏的特色和入门指南，适合新手玩家阅读。",
            tags=["帕鲁", "入门指南", "游戏攻略"],
            category=request.category if request.category else "游戏攻略",
            word_count=len(content),
            model_used="mock-ai-model",
            tokens_used=300,
            generation_time=1.0
        )
    
    async def SearchSimilar(self, request, context):
        """向量搜索"""
        logger.info(f"🔍 收到搜索请求: {request.query}")
        
        # 模拟搜索时间
        await asyncio.sleep(0.3)
        
        from ai_service.app.grpc_generated.ai_service_pb2 import SearchResponse, SearchResult
        
        # 模拟搜索结果
        results = [
            SearchResult(
                id="doc_001",
                title="帕鲁基础攻略",
                content="这是关于帕鲁游戏基础玩法的详细介绍...",
                similarity=0.92,
                metadata={"category": "基础攻略", "author": "AI"}
            ),
            SearchResult(
                id="doc_002", 
                title="高级帕鲁培养技巧",
                content="深入解析帕鲁的培养和进化系统...",
                similarity=0.87,
                metadata={"category": "高级攻略", "author": "AI"}
            )
        ]
        
        return SearchResponse(
            results=results,
            total_count=len(results),
            search_time=0.3,
            query_processed=request.query
        )
    
    async def StreamGenerateContent(self, request, context):
        """流式内容生成"""
        logger.info(f"🌊 收到流式生成请求: {request.prompt[:50]}...")
        
        from ai_service.app.grpc_generated.ai_service_pb2 import StreamContentChunk, ChunkType
        
        content = f"关于{request.prompt}的详细介绍：帕鲁是一款独特的生存冒险游戏，玩家可以在开放世界中捕获、培养和战斗各种可爱的生物。游戏提供了丰富的建造系统和深度的战斗机制。"
        
        # 分块发送内容
        chunk_size = 50
        for i in range(0, len(content), chunk_size):
            chunk = content[i:i + chunk_size]
            is_final = (i + chunk_size) >= len(content)
            
            yield StreamContentChunk(
                chunk=chunk,
                type=ChunkType.CONTENT,
                is_final=is_final
            )
            
            # 模拟流式延迟
            if not is_final:
                await asyncio.sleep(0.2)
        
        # 发送最终元数据
        yield StreamContentChunk(
            chunk="",
            type=ChunkType.FINAL,
            is_final=True,
            metadata={
                "model_used": "mock-ai-model",
                "tokens_used": "200",
                "generation_time": "2.0"
            }
        )
    
    async def GetServiceStats(self, request, context):
        """获取服务统计"""
        logger.info("📊 收到统计请求")
        
        from ai_service.app.grpc_generated.ai_service_pb2 import StatsResponse, ModelStats
        from google.protobuf.timestamp_pb2 import Timestamp
        
        uptime = time.time() - self.start_time
        
        # 模拟模型统计
        model_stats = [
            ModelStats(
                provider="mock-provider",
                model_name="mock-ai-model", 
                requests_count=self.request_count,
                success_rate=0.98,
                avg_response_time=1.2,
                total_tokens=1000,
                status="healthy"
            )
        ]
        
        start_timestamp = Timestamp()
        start_timestamp.FromSeconds(int(self.start_time))
        
        return StatsResponse(
            uptime=uptime,
            total_requests=self.request_count,
            success_requests=int(self.request_count * 0.98),
            error_requests=int(self.request_count * 0.02),
            avg_response_time=1.2,
            models=model_stats,
            memory_usage=0.5,
            vector_db_size=1000,
            start_time=start_timestamp
        )
    
    async def HealthCheck(self, request, context):
        """健康检查"""
        logger.debug("💓 收到健康检查请求")
        
        from ai_service.app.grpc_generated.ai_service_pb2 import HealthResponse, HealthStatus, ComponentHealth
        from google.protobuf.timestamp_pb2 import Timestamp
        
        # 模拟组件健康状态
        components = []
        
        ai_component = ComponentHealth(
            name="mock_ai_service",
            status=HealthStatus.SERVING,
            message="模拟AI服务正常运行"
        )
        ai_component.last_check.GetCurrentTime()
        components.append(ai_component)
        
        response = HealthResponse(
            status=HealthStatus.SERVING,
            service_name="mock-palu-wiki-ai-service",
            version="1.0.0-test",
            components=components,
            message="测试服务正常运行"
        )
        
        response.timestamp.GetCurrentTime()
        return response


async def create_mock_grpc_server(port: int = 50051):
    """创建模拟gRPC服务器"""
    
    # 创建gRPC服务器
    server = aio.server()
    
    # 注册服务
    mock_service = MockAIServiceGRPC()
    
    try:
        from ai_service.app.grpc_generated import ai_service_pb2_grpc
        ai_service_pb2_grpc.add_AIServiceServicer_to_server(mock_service, server)
    except ImportError:
        logger.error("无法导入gRPC服务定义，请确保proto文件已正确生成")
        return None
    
    # 绑定端口
    listen_addr = f'[::]:{port}'
    server.add_insecure_port(listen_addr)
    
    logger.info(f"🚀 模拟gRPC服务器启动在端口 {port}")
    
    # 启动服务器
    await server.start()
    
    return server


async def main():
    """主函数"""
    logger.info("🧪 启动模拟gRPC AI服务器用于测试")
    
    # 启动服务器
    server = await create_mock_grpc_server(50051)
    
    if server:
        logger.info("✅ 模拟gRPC服务器启动成功，等待连接...")
        
        try:
            # 等待服务器终止信号
            await server.wait_for_termination()
        except KeyboardInterrupt:
            logger.info("收到中断信号，正在关闭服务器...")
            await server.stop(grace=5)
    else:
        logger.error("❌ 服务器启动失败")


if __name__ == "__main__":
    asyncio.run(main())