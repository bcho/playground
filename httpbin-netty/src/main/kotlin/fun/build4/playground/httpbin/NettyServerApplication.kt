package `fun`.build4.playground.httpbin

import com.google.gson.Gson
import io.netty.bootstrap.ServerBootstrap
import io.netty.buffer.Unpooled
import io.netty.channel.*
import io.netty.channel.nio.NioEventLoopGroup
import io.netty.channel.socket.SocketChannel
import io.netty.channel.socket.nio.NioServerSocketChannel
import io.netty.handler.codec.http.*
import io.netty.handler.codec.http.HttpHeaderNames.*
import io.netty.handler.codec.http.HttpHeaderValues.CLOSE
import mu.KotlinLogging
import java.net.InetSocketAddress
import java.nio.charset.StandardCharsets

private val logger = KotlinLogging.logger {}
private val gson = Gson()

const val CODEC_HANDLER: String = "handler_codec"
const val HTTP_HANDLER: String = "handler_http"
const val HTTP_RESPONSE_CONTENT_TYPE: String = "application/json"
const val HTTP_HEADER_X_FORWARDED_FOR = "X-Forwarded-For"

class NettyServerApplication(val port: Int) {

    /**
     * Start the server.
     */
    fun start() {
        val bossGroup = NioEventLoopGroup()
        val workerGroup = NioEventLoopGroup()
        try {
            val serverBootstrap = ServerBootstrap()
            serverBootstrap.group(bossGroup, workerGroup)
                .channel(NioServerSocketChannel::class.java)
                .childHandler(object : ChannelInitializer<SocketChannel>() {
                    override fun initChannel(ch: SocketChannel?) {
                        logger.info("received channel")

                        ch!!.pipeline().addLast(CODEC_HANDLER, HttpServerCodec())
                            .addLast(HttpObjectAggregator(10 * 1024 * 1024)) // 10Mb
                            .addLast(HTTP_HANDLER, HttpHandler())
                    }
                })
                .option(ChannelOption.SO_BACKLOG, 128)
                .childOption(ChannelOption.SO_KEEPALIVE, true)

            // bind & start
            val fut = serverBootstrap.bind(port).sync()
            logger.info("server started at :{}", port)

            // wait until the server is closed
            fut.channel().closeFuture().sync()
        } finally {
            workerGroup.shutdownGracefully()
            bossGroup.shutdownGracefully()
        }
    }
}

fun main(args: Array<String>) {
    var port = 8080
    if (args.size > 0) {
        port = Integer.parseInt(args[0])
    }

    NettyServerApplication(port).start()
}

class HttpHandler : ChannelInboundHandlerAdapter() {

    override fun channelReadComplete(ctx: ChannelHandlerContext?) {
        ctx?.flush()
    }

    override fun channelRead(ctx: ChannelHandlerContext, msg: Any?) {
        if (msg == null) {
            return
        }

        logger.info("request started")

        val httpReq = msg as HttpRequest
        val resp = buildAnythingResponseFromRequest(ctx, httpReq)
        sendResponse(ctx, httpReq, resp)

        logger.info("request finished")
    }

    private fun sendResponse(
        ctx: ChannelHandlerContext,
        httpReq: HttpRequest,
        resp: AnythingResponse
    ) {
        val content = gson.toJson(resp).toByteArray(StandardCharsets.UTF_8)

        val httpResp = DefaultFullHttpResponse(
            httpReq.protocolVersion(),
            HttpResponseStatus.OK,
            Unpooled.wrappedBuffer(content)
        )
        httpResp.headers()
            .set(CONTENT_TYPE, HTTP_RESPONSE_CONTENT_TYPE)
            .setInt(CONTENT_LENGTH, httpResp.content().readableBytes())
            .set(CONNECTION, CLOSE)

        val writeFut = ctx.write(httpResp)
        writeFut.addListener { ChannelFutureListener.CLOSE }
    }

    private fun buildAnythingResponseFromRequest(
        ctx: ChannelHandlerContext,
        req: HttpRequest
    ): AnythingResponse {
        return AnythingResponse(
            url = req.uri(),
            method = req.method().toString(),
            origin = getOriginFromRequest(ctx, req),
            headers = getHeadersFromRequest(req)
        )
    }

    private fun getOriginFromRequest(
        ctx: ChannelHandlerContext,
        req: HttpRequest
    ): String {
        if (req.headers().contains(HTTP_HEADER_X_FORWARDED_FOR)) {
            return req.headers().get(HTTP_HEADER_X_FORWARDED_FOR)
        }

        return (ctx.channel().remoteAddress() as InetSocketAddress).address.hostAddress
    }

    private fun getHeadersFromRequest(req: HttpRequest): Map<String, String> {
        return req.headers().entries().map { it.key to it.value }.toMap()
    }
}

data class AnythingResponse(
    val url: String,
    val method: String,
    val origin: String,
    val headers: Map<String, String>
)