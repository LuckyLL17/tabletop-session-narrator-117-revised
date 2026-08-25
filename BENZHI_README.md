# tabletop-session-narrator-117-revised

基于 Go 实现的桌游局势与战报工作台 Web 项目，一款后端服务，这是一个面向桌游爱好者的 Go 全栈记录工具，它不负责玩某一款具体桌游，而是把一局桌游从准备、开始、暂停、回合推进、行动记录到结束复盘的过程保存成可重放的时间线，并生成一份能解释“哪里发生了转折”的战报。

项目源代码、依赖描述和评测专用 Docker 文件共同构成自包含任务；不依赖本机预编译二进制。

## 标准构建、运行和测试命令

```bash
go build ./...
go run ./cmd/server
go test ./...
```
## 评测容器

评测专用 Dockerfile 为 `benzhi.Dockerfile`，构建脚本为 `build_benzhi_docker.sh`。

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh my-go-task linux/arm64
./build_benzhi_docker.sh my-go-task linux/amd64
docker run -it my-go-task:latest
```
