### 常见的信号处理方式

- **优雅关闭**：收到信号后，可以执行一些清理工作，如关闭文件、释放数据库连接等。
- **强制退出**：对于无法忽略的信号（如 `SIGKILL`），通常程序无法进行处理，需要强制终止。

## Signal信号表

| 取值 | **名称**  | **解释**                                                     | **默认动作**           |
| ---- | --------- | ------------------------------------------------------------ | ---------------------- |
| 1    | SIGHUP    | 挂起（在用户终端连接(正常或非正常)结束时发出, 通常是在终端的控制进程结束时, 通知同一session内的各个作业, 这时它们与控制终端不再关联） |                        |
| 2    | SIGINT    | 中断（程序终止(interrupt)信号, 在用户键入INTR字符(通常是Ctrl-C)时发出，用于通知前台进程组终止进程） |                        |
| 3    | SIGQUIT   | 退出（和SIGINT类似, 但由QUIT字符(通常是Ctrl-/)来控制. 进程在因收到SIGQUIT退出时会产生core文件, 在这个意义上类似于一个程序错误信号） |                        |
| 4    | SIGILL    | 非法指令（执行了非法指令. 通常是因为可执行文件本身出现错误, 或者试图执行数据段. 堆栈溢出时也有可能产生这个信号） |                        |
| 5    | SIGTRAP   | 断点或陷阱指令（由断点指令或其它trap指令产生. 由debugger使用） |                        |
| 6    | SIGABRT   | abort发出的信号（调用abort函数生成的信号）                   |                        |
| 7    | SIGBUS    | 非法内存访问（非法地址, 包括内存地址对齐(alignment)出错。比如访问一个四个字长的整数, 但其地址不是4的倍数。它与SIGSEGV的区别在于后者是由于对合法存储地址的非法访问触发的(如访问不属于自己存储空间或只读存储空间)） |                        |
| 8    | SIGFPE    | 浮点异常（在发生致命的算术运算错误时发出. 不仅包括浮点运算错误, 还包括溢出及除数为0等其它所有的算术的错误） |                        |
| 9    | SIGKILL   | kill信号（用来立即结束程序的运行）                           | 不能被忽略、处理和阻塞 |
| 10   | SIGUSR1   | 用户信号1（留给用户使用）                                    |                        |
| 11   | SIGSEGV   | 无效内存访问（试图访问未分配给自己的内存, 或试图往没有写权限的内存地址写数据） |                        |
| 12   | SIGUSR2   | 用户信号2（留给用户使用）                                    |                        |
| 13   | SIGPIPE   | 管道破损，没有读端的管道写数据（这个信号通常在进程间通信产生，比如采用FIFO(管道)通信的两个进程，读管道没打开或者意外终止还往管道写，写进程会收到SIGPIPE信号。此外用Socket通信的两个进程，写进程在写Socket的时候，读进程已经终止，也会产生这个信号） |                        |
| 14   | SIGALRM   | alarm发出的信号（时钟定时信号, 计算的是实际的时间或时钟时间. alarm函数使用该信号） |                        |
| 15   | SIGTERM   | 终止信号（程序结束(terminate)信号, 与SIGKILL不同的是该信号可以被阻塞和处理。通常用来要求程序自己正常退出，shell命令kill缺省产生这个信号） |                        |
| 16   | SIGSTKFLT | 栈溢出                                                       |                        |
| 17   | SIGCHLD   | 子进程退出（子进程结束时, 父进程会收到这个信号）             | 默认忽略               |
| 18   | SIGCONT   | 进程继续                                                     | 不能被阻塞             |
| 19   | SIGSTOP   | 进程停止（停止(stopped)进程的执行. 注意它和terminate以及interrupt的区别:该进程还未结束, 只是暂停执行） | 不能被忽略、处理和阻塞 |
| 20   | SIGTSTP   | 进程停止（停止进程的运行, 用户键入SUSP字符时(通常是Ctrl-Z)发出这个信号） | 该信号可以被处理和忽略 |
| 21   | SIGTTIN   | 进程停止，后台进程从终端读数据时                             |                        |
| 22   | SIGTTOU   | 进程停止，后台进程想终端写数据时                             |                        |
| 23   | SIGURG    | I/O有紧急数据到达当前进程                                    | 默认忽略               |
| 24   | SIGXCPU   | 进程的CPU时间片到期                                          |                        |
| 25   | SIGXFSZ   | 文件大小的超出上限                                           |                        |
| 26   | SIGVTALRM | 虚拟时钟超时                                                 |                        |
| 27   | SIGPROF   | profile时钟超时                                              |                        |
| 28   | SIGWINCH  | 窗口大小改变                                                 | 默认忽略               |
| 29   | SIGIO     | I/O相关                                                      |                        |
| 30   | SIGPWR    | 关机                                                         | 默认忽略               |
| 31   | SIGSYS    | 系统调用异常                                                 |                        |



### 常用的信号类型

1. `SIGINT`：通常由用户在终端按 `Ctrl+C` 触发，表示程序中断。
2. `SIGTERM`：请求程序正常终止。
3. `SIGHUP`：终端关闭或配置文件发生更改。
4. `SIGQUIT`：程序异常退出，通常由用户按 `Ctrl+\` 触发。
5. `SIGKILL`：强制杀死进程，无法捕获或忽略。
6. `SIGSTOP`：暂停进程，不能捕获或忽略。

#####  使用 `os/signal` 包处理信号

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 创建一个信号通道
	signalChan := make(chan os.Signal, 1)
	
	// 捕获 SIGINT 和 SIGTERM 信号
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动一个协程模拟工作
	go func() {
		for {
			fmt.Println("程序运行中...")
			time.Sleep(1 * time.Second)
		}
	}()

	// 等待信号到达
	sigReceived := <-signalChan
	fmt.Println("收到信号:", sigReceived)

	// 做一些清理工作
	fmt.Println("正在清理资源...")
	time.Sleep(2 * time.Second)

	// 优雅退出
	fmt.Println("程序退出")
}

```

