# 第一阶段：使用Node镜像构建前端
FROM node:latest AS node_builder
WORKDIR /app
COPY front/ ./front/
RUN cd front && yarn install && yarn build

# 第二阶段：使用Go镜像编译后端
FROM golang:1.22 AS go_builder
WORKDIR /app
COPY . .
COPY --from=node_builder /app/front/dist ./front/dist
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o aliyun .

# 第三阶段：设置最终的运行环境
FROM alpine:latest
WORKDIR /app
RUN mkdir -p /etc/aliyun
COPY --from=go_builder /app/aliyun ./
ENTRYPOINT ["/app/aliyun"]