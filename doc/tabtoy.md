# 					   TABTOY

高性能表格数据导出器

项目地址：https://github.com/davyxu/tabtoy

[toc]

## 基础表

* 类型表 Type.xlsx 名称一般不改，也可以自定义， 顾名思义，就是定义类型的表
* 数据表 ***.xlsx 自定义名称即可，就是我们的数据表,可以有很多个
* 索引表 Index.xlsx 名称一般不改，也可以自定义，将所有的需要导出的表名放进来



## 一、``` -mode ``` 版本有三个版本 

## 	v1,v2,v3  最新的v3  所以我们就研究v3即可



## 二、```-index```=Index.xlsx 

## 	索引表地址



## 三、```-go_out```=table_gen.go 

## 	生成的go文件地址  -package=main 生成的golang的包名



## 四、```-json_out```=table_gen.json 

## 	生成的json文件和地址



## 五、```-proto_out```=table.proto 

## 	生成的proto文件即地址

## 	```-pbbin_out```=all.pbb 导出proto二进制文件

### protoc --go_out=. ./table.proto -I . 将proto文件生成代码	



## 六、```--tag_action```=action1:tag1+tag2|action2:tag1+tag3
```shell script
--tag_action=action1:tag1+tag2|action2:tag1+tag3
```
* | 表示多个action
* 被标记的tag, 将被对应action处理

### action类型
action | 适用范围 | 功能
---|---|---|
nogenfield_json | Type表 | 被标记的字段不导出到json完整文件中
nogenfield_jsondir| Type表 | 被标记的字段不导出到每个表文件json
nogenfield_binary| Type表 | 被标记的字段不导出到二进制中
nogenfield_pbbin| Type表 | 被标记的字段不导出到Protobuf二进制中
nogenfield_lua| Type表 | 被标记的字段不导出到Lua中
nogenfield_csharp| Type表 | 被标记的字段不导出到C#中
nogentab| Index表 | 被标记的表不会导出到任何输出中

## 七、```-json_dir```=导出到那个目录下 按需读取

## 八、```-usecache```=true 是否启用缓存功能
-cachedir参数设定缓存目录, 默认缓存到tabtoy当前目录下的.tabtoycache目录

## 注：也可以导出为C#,java,lua,json等类型的

## 类型表 Type.xlsx
种类 | 对象类型 | 标识名 | 字段名 | 字段类型 | 数组切割| 值 | 索引 | 标记
---|---|---|---|---|---|---|---|---

#### 种类
* 枚举
* 表头

#### 对象类型
* 枚举的值 就填定义的枚举类型 例如 AchievementType
* 表头的值 表名（一般为表名，也可以更改）

#### 标志名
* 枚举第一个 一般为空 其余的填表中的枚举值的名称，例如 活跃成就
* 表头的就填表头的名称

#### 字段名
* 生成json，go文件中的名称，一般为英文名

#### 字段类型
* 生成的字段名的类型 int，string，float，double，等，具体的枚举的值使用枚举的对象类型

#### 数组切割
* 数组类型的字段必填|，数组类型的表，可以在数据表中定义两个相同的标识名，或者在数据表的数据中用|分割

#### 值
* 枚举类型填写枚举之后的值

#### 索引
* 使用索引的填写是，不使用索引的为空即可

#### 标记
* 标记功能

## 数据表 ***.xlsx
一般第一行为数据名，其余行都为数据


## 索引表 Index.xlsx
模式 | 表类型 | 表文件名 | 标记 
---|---|---|---

#### 模式
* 类型表 定义的类型表放进来
* 数据表 其余的数据表放进来
* 键值表 只存在键值对的数据表

#### 表类型 
* 就是我们代码中结构体的名称

#### 表文件名
* 文件名

#### 标记
* 标记功能

## 键值表 
字段名 | 字段类型	| 标识名	| 值	| 数组切割	| 标记
---|---|---|---|---|---

键值表数据不需要写入Type表



## 表拆分

将ExampleData表, 拆为Data.csv, Data2.csv表

模式 | 表类型 | 表文件名
---|---|---
类型表 |        | Type.xlsx
数据表 | ExampleData | Data.csv
数据表 | ExampleData | Data2.csv

每个表中的字段可按需填写


## 空行分割

表格数据如下:

ID | 名称
---|---
1 | 坦克
2 | 法师
(空行)  |
3 | 治疗

导出数据
* 1 坦克
* 2 法师

导表工具在识别到空行后, 空行后的数据将被忽略


## 行数据注释

表格数据如下:

ID | 名称
---|---
1 | 坦克
#2 | 法师
3 | 治疗

导出数据
* 1 坦克
* 3 治疗

在任意表的首列单元格中首字符为#时，该行所有数据不会被导出

## 列数据注释

表格数据如下:

ID | #名称
---|---
1 | 坦克
2 | 法师
3 | 治疗

导出数据
* 1
* 2
* 3

表头中, 列字段首字符为#时，该列所有数据按默认值导出 



