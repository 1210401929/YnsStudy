#!/bin/bash
# 在容器 NS 里重启 Go 服务：停掉旧进程，等 8889 端口空出来，带上服务器专用的地址重新启动。
# 用法（在宿主机上执行）：docker exec NS bash /opt/Go/restart.sh
#
# 服务器上 MySQL 在宿主机、Redis 在容器里，不用绕公网：
#   NYS_DATABASE_HOST=172.17.0.1  Docker 默认网桥的网关，就是宿主机
#   NYS_REDIS_HOST=127.0.0.1      Redis 和 Go 在同一个容器里
# 本地开发不用这个脚本，仍按 config.yaml 连远程。

cd /opt/Go || exit 1

if pgrep -f "./nys-go-api" > /dev/null; then
    echo "停止旧进程..."
    pkill -f "./nys-go-api"
    for i in $(seq 1 15); do
        pgrep -f "./nys-go-api" > /dev/null || break
        sleep 1
    done
    if pgrep -f "./nys-go-api" > /dev/null; then
        echo "旧进程 15 秒内没有退出，请检查后再试" >&2
        exit 1
    fi
fi

chmod +x ./nys-go-api
NYS_DATABASE_HOST=172.17.0.1 NYS_REDIS_HOST=127.0.0.1 nohup ./nys-go-api > nohup.out 2>&1 &

sleep 3
echo "----- 启动日志 -----"
tail -n 20 nohup.out
if pgrep -f "./nys-go-api" > /dev/null; then
    echo "----- 已启动 -----"
else
    echo "----- 启动失败，请查看上面的日志 -----" >&2
    exit 1
fi
