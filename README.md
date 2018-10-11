# Golang Conn
使用[netpoll](https://github.com/mailru/easygo)将连接处理改为事件驱动，并建立goroutine协程池化，使协程数<连接数，具体降到多少根据数据传输的频次评估，尤其对于高连接数、低频数据传输的场景，可以大幅降低协程数与连接数的比例。

- 事件驱动模型

![conn](/doc/img/conn.jpg "conn")

## 参考内容
- [[转][译]百万级WebSockets和Go语言](https://colobu.com/2017/12/13/A-Million-WebSockets-and-Go/)
    - [[原]A Million WebSockets and Go](https://medium.freecodecamp.org/million-websockets-and-go-cc58418460bb)