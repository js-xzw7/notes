#### rabbitMQ 

rabbitMQ 是一个消息代理：它接收并转发消息。你可以把它想象成一个邮局：当你把要寄出的邮件放入邮箱时，邮递员最终会将邮件投递给收件人。在这个比喻中，rabbitMQ既是邮局，同时也是邮递员。

rabbitMQ与邮局之间的主要区别在于它不处理纸张，而是接受、存储和转发二进制数据块（即消息）。

rabbitMQ术语

- 生产者（producer）的意思就是发送消息。发送消息的程序就是生产者。

- 队列（queue）是rabbitMQ中邮箱的名称。尽管消息在rabbitMQ和您的应用程序中流动，但它们只能存储在队列中。队列仅受主机内存和磁盘限制，它本质上是一个大型消息缓冲区。

  多个生产者可以发送消息到一个队列，多个消费者可以尝试从一个队列接收数据。

- 消费者（consumer）与接收者含义类似。消费者是一个主要等待接收消息的程序。



#### 模式

##### 1.简单模式

*一对一，一生产者， 一个消费者。*

###### 示例

生产者：

```go
package main

import (
    "context"
    amqp "github.com/rabbitmq/amqp091-go"
    "log"
    "time"
)

func failOnError(err error, msg string) {
    if err != nil {
       log.Panicf("%s:%s", msg, err)
    }
}
func main() {
    //1.连接rabbitmq服务器
    conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
    failOnError(err, "Failed to connect to rabbitmq")
    defer conn.Close()

    //2.创建通道，大部分用于完成操作的api都驻留在通道中
    ch, err := conn.Channel()
    failOnError(err, "failed to open a channel")
    defer ch.Close()

    //3.声明要发送的队列,声明队列是幂等的，只有当队列不存在时才会创建
    q, err := ch.QueueDeclare(
       "hello",
       false,
       false,
       false,
       false,
       nil,
    )
    failOnError(err, "failed to declare a queue")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    body := "Hello World!"
    err = ch.PublishWithContext(
       ctx,
       "",
       q.Name,
       false,
       false,
       amqp.Publishing{
          ContentType: "text/plain",
          Body:        []byte(body),
       })

    failOnError(err, "failed to publish a message")
    log.Printf("[x] sent %s \n", body)
}
```

消费者：

```go
package main

import (
    amqp "github.com/rabbitmq/amqp091-go"
    "log"
)

func failOnError(err error, msg string) {
    if err != nil {
       log.Panicf("%s:%s", msg, err)
    }
}

func main() {
    conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
    failOnError(err, "failed to connect to rabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
       "hello",
       false,
       false,
       false,
       false,
       nil,
    )

    failOnError(err, "failed to declare a queue")

    msgs, err := ch.Consume(
       q.Name,
       "",
       true,
       false,
       false,
       false,
       nil,
    )
    failOnError(err, "failed to register a consumer")

    var forever chan struct{}
    go func() {
       for d := range msgs {
          log.Printf("received a message:%s", d.Body)
       }
    }()

    log.Printf("[*] waiting for messages, to exit press ctrl+c")
    <-forever
}
```

##### 2.工作队列

*一对多，一个生产者，多个消费者。*用于多个工作线程之间分配耗时任务。

工作队列（又称任务队列）的核心思想是避免立即执行资源密集型任务并等待其完成。相反，我们将任务安排到稍后执行。我们将任务封装为消息并将其发送到队列。后台运行的工作进程会弹出任务并最终执行该作业。当运行多个工作进程时，任务将在它们之间共享。

###### 循环

使用任务队列的优势之一是能够轻松地并行处理工作。如果我们积压了大量工作，只需要添加更多的工作线程即可，从而轻松实现扩展。默认情况下rabbitMQ会**按照顺序** 将每条消息发送给下一个消费者。平均而言，每个消费者会受到相同数量的消息。这一种消息分发方式成为轮询。

###### 消息确认

执行一项任务可能需要几秒钟，您可能想知道如果消费者启动了一个耗时较长的任务，并且在完成之前终止会发生什么。未启动消息确认模式，rabbitMQ一旦向消费者发送一条消息，就会立即将其标记为删除。在这种情况下，如果您终止一个工作进程，它刚刚处理的消息就会丢失。已发送给该特定工作进程但尚未处理的消息也会丢失。

为了确保消息永不丢失，rabbitMQ支持消息确认。消费者会发送一个确认消息，告知rabbitMQ某条消息已被接收处理，并且可以删除该消息。

如果某个消息挂掉（其通道关闭、连接关闭或者tcp连接丢失）且未发送确认消息，rabbitMQ会认为该消息未得到完全处理，并将其重新放入队列。如果同时有其他消费者在线，它会快速将该消息重启投递给其他消费者。这样，即使工作线程偶尔挂掉，也能确保消息不会丢失。

消费者的送达确认时间强制设置超时时间（默认为30分钟），这有助于检测那些始终未确认送达的、存在问题的（卡住的）消费者。可自行配置超时时间。

###### 消息持久性

如果rabbitMQ退出或崩溃时，它会忘记队列和消息。未了确保消息不丢失，需要做两件事情：将队列和消息都标记为持久化。

**将消息标记为持久化并不能完全保证消息不会丢失。**虽然它告诉 RabbitMQ 将消息保存到磁盘，但 RabbitMQ 仍然会在短时间内接收消息但尚未保存。此外，RabbitMQ 并非`fsync(2)`对每条消息都这样做——它可能只是被保存到缓存中，而并未真正写入磁盘。持久化保证并不强，但对于我们这个简单的任务队列来说已经足够了。如果您需要更强的保证，可以使用 [发布者确认 (publisher confirmed)](https://www.rabbitmq.com/docs/confirms)。

###### 公平调度

在有两个 Worker 的情况下，如果奇数消息较多，偶数消息较少，那么其中一个 Worker 就会一直处于忙碌状态，而另一个 Worker 几乎不做任何工作。然而，RabbitMQ 对此一无所知，仍然会均匀地分发消息。

发生这种情况的原因是，RabbitMQ 只是在消息进入队列时才发送消息。它不会查看消费者未确认消息的数量。它只是盲目地将每第 n 条消息发送给第 n 个消费者。

为了解决这个问题，我们可以将预取计数设置为`1`。这将告诉 RabbitMQ 不要一次向一个 Worker 发送多条消息。或者换句话说，在 Worker 处理并确认上一条消息之前，不要向它发送新消息。相反，它会将新消息发送给下一个仍处于空闲状态的 Worker。

###### 示例

生产者

```go
package main

import (
    "context"
    amqp "github.com/rabbitmq/amqp091-go"
    "log"
    "os"
    "strings"
    "time"
)

func failOnError(err error, msg string) {
    if err != nil {
       log.Panicf("%s:%s", msg, err)
    }
}

func bodyFrom(args []string) string {
    var s string
    if (len(args) < 2) || os.Args[1] == "" {
       s = "hello"
    } else {
       s = strings.Join(args[1:], " ")
    }
    return s
}
func main() {
    //1.连接rabbitmq服务器
    conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
    failOnError(err, "Failed to connect to rabbitmq")
    defer conn.Close()

    //2.创建通道，大部分用于完成操作的api都驻留在通道中
    ch, err := conn.Channel()
    failOnError(err, "failed to open a channel")
    defer ch.Close()

    //3.声明要发送的队列,声明队列是幂等的，只有当队列不存在时才会创建
    q, err := ch.QueueDeclare(
       "hello",
       false,
       false,
       false,
       false,
       nil,
    )
    failOnError(err, "failed to declare a queue")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    body := bodyFrom(os.Args)
    err = ch.PublishWithContext(
       ctx,
       "",
       q.Name,
       false,
       false,
       amqp.Publishing{
          DeliveryMode: amqp.Persistent,
          ContentType:  "text/plain",
          Body:         []byte(body),
       })

    failOnError(err, "failed to publish a message")
    log.Printf("[x] sent %s \n", body)
}
```

消费者

```go
package main

import (
    "bytes"
    amqp "github.com/rabbitmq/amqp091-go"
    "log"
    "time"
)

func failOnError(err error, msg string) {
    if err != nil {
       log.Panicf("%s:%s", msg, err)
    }
}

func main() {
    conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
    failOnError(err, "failed to connect to rabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
       "hello",
       false,
       false,
       false,
       false,
       nil,
    )

    failOnError(err, "failed to declare a queue")

    msgs, err := ch.Consume(
       q.Name,
       "",
       false,
       false,
       false,
       false,
       nil,
    )
    failOnError(err, "failed to register a consumer")

    var forever chan struct{}
    go func() {
       for d := range msgs {
          log.Printf("received a message:%s", d.Body)
          dotCount := bytes.Count(d.Body, []byte("."))
          t := time.Duration(dotCount)
          time.Sleep(t * time.Second)
          log.Printf("Done")
          d.Ack(false)
       }
    }()

    log.Printf("[*] waiting for messages, to exit press ctrl+c")
    <-forever
}
```



##### 3.发布/订阅

*一对多* 

工作队列的假设是每个任务只会传递给一个工作线程，如果将一条消息传递给多个消费者，这种模式被称为“发布/订阅”

rabbitMQ消息模型的核心思想是**生产者永远不会直接向队列发送消息**。实际上，生产者很多时候甚至不知道消息是否会被投递到任何队列。

###### 交换机

生产者只能将消息发送到**交换机**。交换机非常简单，它一方面接收来来自生产者的消息，另一方面将消息推送到消息队列。

######  交换机类型

交换机必须确切地知道如果处理收到的消息。应该将其添加到特定的队列吗？应该将其添加到多个队列吗？还是应该放弃？这些规则由交换机类型定义，可用交换机类型：direct、topic、headers、fanout。

默认交换机：空字符串（""）标识；如果不声明交换机，使用“”标识，也是能够向队列发送消息的

###### 临时队列

当我们声明一个队列时，提供的队列名为空字符串时，会创建一个具有生成名称的非持久队列。一旦消费者断开连接，队列自动删除。如果我们只是对当前正在流转的的消息感兴趣，而不是旧的消息（比如日志消息），临时队列对我们很有用

###### 示例

生产者：

```go
package main

import (
    "context"
    amqp "github.com/rabbitmq/amqp091-go"
    "log"
    "os"
    "strings"
    "time"
)

func failOnError(err error, msg string) {
    if err != nil {
       log.Panicf("%s:%s", msg, err)
    }
}

func bodyFrom(args []string) string {
    var s string
    if (len(args) < 2) || os.Args[1] == "" {
       s = "hello"
    } else {
       s = strings.Join(args[1:], " ")
    }
    return s
}
func main() {
    //1.连接rabbitmq服务器
    conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
    failOnError(err, "Failed to connect to rabbitmq")
    defer conn.Close()

    //2.创建通道，大部分用于完成操作的api都驻留在通道中
    ch, err := conn.Channel()
    failOnError(err, "failed to open a channel")
    defer ch.Close()

    //3.声明交换机
    err = ch.ExchangeDeclare(
       "logs",
       "fanout",
       true,
       false,
       false,
       false,
       nil,
    )

    failOnError(err, "failed to declare an exchange")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    body := bodyFrom(os.Args)
    err = ch.PublishWithContext(
       ctx,
       "logs",
       "",
       false,
       false,
       amqp.Publishing{
          DeliveryMode: amqp.Persistent,
          ContentType:  "text/plain",
          Body:         []byte(body),
       })

    failOnError(err, "failed to publish a message")
    log.Printf("[x] sent %s \n", body)
}
```

消费者：

```go
package main

import (
    amqp "github.com/rabbitmq/amqp091-go"
    "log"
)

func failOnError(err error, msg string) {
    if err != nil {
       log.Panicf("%s:%s", msg, err)
    }
}

func main() {
    //1.建立连接
    conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
    failOnError(err, "failed to connect to rabbitMQ")
    defer conn.Close()

    //2.建立通道
    ch, err := conn.Channel()
    failOnError(err, "failed to open a channel")
    defer ch.Close()

    //3.声明交换机
    err = ch.ExchangeDeclare(
       "logs",
       "fanout",
       true,
       false,
       false,
       false,
       nil,
    )

    failOnError(err, "failed to declare an exchange")

    //4.声明队列
    q, err := ch.QueueDeclare(
       "",
       false,
       false,
       false,
       false,
       nil,
    )
    failOnError(err, "failed to declare a queue")

    //5.绑定队列
    err = ch.QueueBind(
       q.Name,
       "",
       "logs",
       false,
       nil)

    failOnError(err, "failed to bind a queue")

    //6.消费队列
    msgs, err := ch.Consume(
       q.Name,
       "",
       false,
       false,
       false,
       false,
       nil,
    )
    failOnError(err, "failed to register a consumer")

    var forever chan struct{}
    go func() {
       for d := range msgs {
          log.Printf(" [x] %s", d.Body)

       }
    }()

    log.Printf("[*] waiting for messages, to exit press ctrl+c")
    <-forever
}
```

##### 4.路由

订阅消息队列中的子集。比如日志消息仅将关键错误信息定向到日志文件

###### 示例

生产者：

```go
package main

import (
    "context"
    amqp "github.com/rabbitmq/amqp091-go"
    "log"
    "os"
    "strings"
    "time"
)

func failOnError(err error, msg string) {
    if err != nil {
       log.Panicf("%s:%s", msg, err)
    }
}

func bodyFrom(args []string) string {
    var s string
    if (len(args) < 2) || os.Args[1] == "" {
       s = "hello"
    } else {
       s = strings.Join(args[1:], " ")
    }
    return s
}

func severityFrom(args []string) string {
    var s string
    if (len(args) < 2) || os.Args[1] == "" {
       s = "info"
    } else {
       s = os.Args[1]
    }

    return s
}
func main() {
    //1.连接rabbitmq服务器
    conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
    failOnError(err, "Failed to connect to rabbitmq")
    defer conn.Close()

    //2.创建通道，大部分用于完成操作的api都驻留在通道中
    ch, err := conn.Channel()
    failOnError(err, "failed to open a channel")
    defer ch.Close()

    //3.声明交换机
    err = ch.ExchangeDeclare(
       "logs_direct",
       "direct",
       true,
       false,
       false,
       false,
       nil,
    )

    failOnError(err, "failed to declare an exchange")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    body := bodyFrom(os.Args)
    err = ch.PublishWithContext(
       ctx,
       "logs_direct",
       severityFrom(os.Args),
       false,
       false,
       amqp.Publishing{
          DeliveryMode: amqp.Persistent,
          ContentType:  "text/plain",
          Body:         []byte(body),
       })

    failOnError(err, "failed to publish a message")
    log.Printf("[x] sent %s \n", body)
}
```

消费者：

```go
package main

import (
    amqp "github.com/rabbitmq/amqp091-go"
    "log"
    "os"
)

func failOnError(err error, msg string) {
    if err != nil {
       log.Panicf("%s:%s", msg, err)
    }
}

func main() {
    //1.建立连接
    conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
    failOnError(err, "failed to connect to rabbitMQ")
    defer conn.Close()

    //2.建立通道
    ch, err := conn.Channel()
    failOnError(err, "failed to open a channel")
    defer ch.Close()

    //3.声明交换机
    err = ch.ExchangeDeclare(
       "logs_direct",
       "direct",
       true,
       false,
       false,
       false,
       nil,
    )

    failOnError(err, "failed to declare an exchange")

    //4.声明队列
    q, err := ch.QueueDeclare(
       "",
       false,
       false,
       false,
       false,
       nil,
    )
    failOnError(err, "failed to declare a queue")

    if len(os.Args) < 2 {
       log.Printf("Usage: %s [info] [warning] [error]", os.Args[0])
       os.Exit(0)
    }

    for _, s := range os.Args[1:] {
       log.Printf("Binding queue %s to exchange %s with routing key %s", q.Name, "logs_direct", s)
       //5.绑定队列
       err = ch.QueueBind(
          q.Name,
          s,
          "logs_direct",
          false,
          nil)

       failOnError(err, "failed to bind a queue")
    }

    //6.消费队列
    msgs, err := ch.Consume(
       q.Name,
       "",
       false,
       false,
       false,
       false,
       nil,
    )
    failOnError(err, "failed to register a consumer")

    var forever chan struct{}
    go func() {
       for d := range msgs {
          log.Printf(" [x] %s", d.Body)

       }
    }()

    log.Printf("[*] waiting for messages, to exit press ctrl+c")
    <-forever
}
```

##### 5.主题

 路由模式改进了我们的系统，但它仍然存在局限性 - 它无法根据多个条件进行路由。

发送到 `topic` 交换机的消息不能具有任意的 `routing_key` - 它必须是单词列表，用点分隔。单词可以是任何内容，但通常它们指定与消息相关的某些特征。一些有效的路由键示例：`stock.usd.nyse`, `nyse.vmw`, `quick.orange.rabbit`。路由键中可以包含任意数量的单词，最多 255 字节的限制。

绑定键也必须采用相同的形式。`topic` 交换机背后的逻辑类似于 `direct` 交换机 - 使用特定路由键发送的消息将被传递到所有使用匹配绑定键绑定的队列。但是，绑定键有两个重要的特殊情况

- `*` (星号) 可以替代正好一个单词。
- `#` (井号) 可以替代零个或多个单词。

###### 主题 (Topic) 交换机

主题 (Topic) 交换机功能强大，可以像其他交换机一样工作。

当队列绑定了 `#` (井号) 绑定键时 - 它将接收所有消息，无论路由键如何 - 就像 `fanout` 交换机一样。

当特殊字符 `*` (星号) 和 `#` (井号) 未在绑定中使用时，主题 (topic) 交换机将像 `direct` 交换机一样工作。

###### 示例

生产者：

```go
package main

import (
    "context"
    amqp "github.com/rabbitmq/amqp091-go"
    "log"
    "os"
    "strings"
    "time"
)

func failOnError(err error, msg string) {
    if err != nil {
       log.Panicf("%s:%s", msg, err)
    }
}

func bodyFrom(args []string) string {
    var s string
    if (len(args) < 2) || os.Args[1] == "" {
       s = "hello"
    } else {
       s = strings.Join(args[1:], " ")
    }
    return s
}

func severityFrom(args []string) string {
    var s string
    if (len(args) < 2) || os.Args[1] == "" {
       s = "info"
    } else {
       s = os.Args[1]
    }

    return s
}
func main() {
    //1.连接rabbitmq服务器
    conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
    failOnError(err, "Failed to connect to rabbitmq")
    defer conn.Close()

    //2.创建通道，大部分用于完成操作的api都驻留在通道中
    ch, err := conn.Channel()
    failOnError(err, "failed to open a channel")
    defer ch.Close()

    //3.声明交换机
    err = ch.ExchangeDeclare(
       "logs_topic",
       "topic",
       true,
       false,
       false,
       false,
       nil,
    )

    failOnError(err, "failed to declare an exchange")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    body := bodyFrom(os.Args)
    err = ch.PublishWithContext(
       ctx,
       "logs_topic",
       severityFrom(os.Args),
       false,
       false,
       amqp.Publishing{
          DeliveryMode: amqp.Persistent,
          ContentType:  "text/plain",
          Body:         []byte(body),
       })

    failOnError(err, "failed to publish a message")
    log.Printf("[x] sent %s \n", body)
}
```

消费者：

```go
package main

import (
    amqp "github.com/rabbitmq/amqp091-go"
    "log"
    "os"
)

func failOnError(err error, msg string) {
    if err != nil {
       log.Panicf("%s:%s", msg, err)
    }
}

func main() {
    //1.建立连接
    conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
    failOnError(err, "failed to connect to rabbitMQ")
    defer conn.Close()

    //2.建立通道
    ch, err := conn.Channel()
    failOnError(err, "failed to open a channel")
    defer ch.Close()

    //3.声明交换机
    err = ch.ExchangeDeclare(
       "logs_topic",
       "topic",
       true,
       false,
       false,
       false,
       nil,
    )

    failOnError(err, "failed to declare an exchange")

    //4.声明队列
    q, err := ch.QueueDeclare(
       "",
       false,
       false,
       false,
       false,
       nil,
    )
    failOnError(err, "failed to declare a queue")

    if len(os.Args) < 2 {
       log.Printf("Usage: %s [info] [warning] [error]", os.Args[0])
       os.Exit(0)
    }

    for _, s := range os.Args[1:] {
       log.Printf("Binding queue %s to exchange %s with routing key %s", q.Name, "logs_direct", s)
       //5.绑定队列
       err = ch.QueueBind(
          q.Name,
          s,
          "logs_topic",
          false,
          nil)

       failOnError(err, "failed to bind a queue")
    }

    //6.消费队列
    msgs, err := ch.Consume(
       q.Name,
       "",
       false,
       false,
       false,
       false,
       nil,
    )
    failOnError(err, "failed to register a consumer")

    var forever chan struct{}
    go func() {
       for d := range msgs {
          log.Printf(" [x] %s", d.Body)

       }
    }()

    log.Printf("[*] waiting for messages, to exit press ctrl+c")
    <-forever
}
```
