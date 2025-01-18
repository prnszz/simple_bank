# 学习笔记
学习simple_bank项目的笔记

## lecture 1
使用`dbdiagram.io`创建数据库表结构图, 并且他会自动生成sql语句用来创建表结构

## lecture 2
下载并且使用`Docker`, `TablePlus`
```shell
# 一些基本操作
docker ps # 查看正在运行的容器
docker ps -a # 查看所有容器
# image是容器的模板, container是运行的实例
docker pull postgres # 下载postgres镜像

# docker run --name <container_name> -e <env_variable> -p <host_port>:<container_port> -d <image_name>:<tag>
# details at makefile

# 进入容器
docker exec -it <container_name_or_id> <command> [args]
docker exec -it simple_bank_db psql -U postgres

# logs
docker logs <container_name_or_id>
```

### 补充
```shell
# 删除单个镜像
docker rmi <image_id>
# 删除多个镜像
docker rmi <image_id_1> <image_id_2>

# 只清理未使用的镜像
docker system prune -a

# 同时清理未使用的镜像和构建缓存
docker system prune -a --volumes

# 查看哪些接口被占用
# netstat -an | grep <port>
netstat -an | grep 5432

# shutdown local PostgreSQL
brew services stop postgresql
```

## lecture 3 DB migration
使用migrate CLI工具来管理数据库迁移
```shell
$ brew install golang-migrate

migrate create -ext sql -dir db/migration -seq init_schema
```
运行之后会在`db/migration`目录下生成两个文件, 一个是`up.sql`, 一个是`down.sql`
我们把创建表的sql语句(dbdiagram)写在`up.sql`中, 删除表的sql语句写在`down.sql`中

然后我们把这些命令写到`Makefile`中, 以便于我们使用
详细请见 `migrate up` 和 `migrate down` 的命令 

## lecture 4 CRUD go-sqlc
```shell
brew install sqlc
sqlc version
sqlc init
```
然后我们需要配置`sqlc.yaml`文件, 详细请见`sqlc.yaml`文件
之后我们为sql query根据不同的业务创建不同的`sql`文件, 详细请见`db/`目录下的文件
运行`sqlc generate`, 之后会在`db`目录下生成`db.go`文件, 里面包含了我们的`query`方法, `models.go`文件, 里面包含了我们的`struct`定义

## lecture 5 unit test
我们使用`testify`来进行单元测试
```shell
go get github.com/lib/pq
go get github.com/stretchr/testify
```
1. 测试入口在`db/sqlc/main_test.go`里面. `TestMain`是测试的入口, 会在测试开始之前创建一个数据库连接, 并且在测试结束之后关闭数据库连接
2. 测试工具在`util`目录下, 里面包含了一些测试的工具函数
3.最佳实践是根据不同的`.sql.go`文件创建不同的测试文件.  

## lecture 6 DB transaction
DB transaction(事务)是一个工作单元. 它具有以下四个属性:
1. 原子性(Atomicity): 事务是一个原子操作, 由一系列操作组成. 事务的原子性确保事务中的所有操作要么全部完成, 要么全部不完成
2. 一致性(Consistency): 事务开始之前和事务结束之后, 数据库的完整性约束没有被破坏
3. 隔离性(Isolation): 事务的执行不会受到其他事务的干扰
4. 持久性(Durability): 事务完成之后, 事务对数据库的所有更新将被保存到数据库中

`db/sqlc/store.go`中建立了一个`Store`结构体, 里面包含了一个`*sql.DB`的指针, 用来操作数据库.还有`Queries`结构体, 里面包含了我们的`query`方法

`execTx`方法是一个事务的封装, 里面包含了我们的事务操作. 参数是`context`, `fn func(*Queries) error`, 返回值是`error`. `fn`函数, 里面包含了我们的事务操作. `execTx`方法会在`fn`函数执行之前开启一个事务, 并且在`fn`函数执行之后根据`fn`函数的返回值决定是`commit`还是`rollback`事务

`TransferTx`方法是一个转账实务操作. 他会使用`exeTx`方法来执行实务操作. 

最后我们写一下测试用例`db/sqlc/store_test.go`, 来测试我们的事务操作

## lecture 7 DB TX lock
什么事deadlock? 一个事务在等待另一个事务释放锁, 而另一个事务又在等待第一个事务释放锁. 这样就会导致死锁. 
这一章节的debug思路比较重要. 主要是创建了一个map, 用来查看事务的状态.

在`TransferTx`方法中, 我们使用`SELECT ... FOR UPDATE`来锁定行, 防止并发操作. 但是这样会导致数据库的性能下降, 因为锁定行会导致其他事务无法操作这行数据. 所以我们需要在`SELECT`语句后面加上`SKIP LOCKED`来跳过锁定行, 这样可以提高数据库的性能

## lecture 8 avoid deadlock in transaction
在account1 -> account2的转账操作中, 我们使用了`SELECT ... FOR UPDATE`来锁定行, 避免了这个事务的并发操作时的问题. 但是这并没有避免account2 -> account1的转账操作时的问题. 
account2 -> account1进行转账操作时, 如果account1 -> account2的转账操作还没有完成, 那么account2 -> account1的转账操作就会等待account1 -> account2的转账操作完成, 这样就会导致死锁.
我们的思路还是先修改test case, 然后再修改代码. 
我们主要是需要需要确保所有的事务都是按照相同的顺序来锁定行. 例如: account1 -> account2的转账操作中, 我们先锁定account1, 再锁定account2. 
account2 -> account1的转账操作中, 我们也是先锁定account1, 再锁定account2. 这样就可以避免死锁的问题.

## lecture 9 数据库事务隔离级别详解：MySQL vs PostgreSQL

### 1. 基础概念
事务隔离是ACID特性（原子性、一致性、隔离性、持久性）中的重要组成部分，用于确保并发事务不会相互影响。ANSI标准定义了四个隔离级别，用于解决不同的并发问题。

### 2. 四种读取现象（并发问题）
* **脏读**：一个事务读取到另一个事务尚未提交的数据变更。这种数据可能会被回滚，导致读取到的数据是无效的。
* **不可重复读**：在同一个事务内多次读取同一数据，由于其他事务的提交导致每次读取的结果不同。这破坏了事务内数据的一致性。
* **幻读**：在同一事务内执行相同的查询，由于其他事务插入或删除了符合查询条件的数据，导致查询返回的结果集发生变化。
* **序列化异常**：多个并发事务的执行结果，无法通过将这些事务按某种顺序串行执行得到相同的结果。

### 3. 隔离级别对比

| 特性 | MySQL | PostgreSQL |
|------|--------|------------|
| 默认隔离级别 | Repeatable Read | Read Committed |
| Read Uncommitted支持 | 支持 | 不支持（等同于Read Committed） |
| 实现机制 | 锁机制 | 依赖检测机制 |
| 隔离级别设置范围 | 支持全局和会话级别 | 仅支持事务内部设置 |
| 可重复读级别的并发更新 | 允许某些不一致的并发更新 | 更严格，会抛出错误 |

### 4. 各隔离级别防止现象对比

| 隔离级别 | 脏读 | 不可重复读 | 幻读 | 序列化异常 |
|---------|------|------------|------|------------|
| MySQL Read Uncommitted | ❌ | ❌ | ❌ | ❌ |
| MySQL Read Committed | ✅ | ❌ | ❌ | ❌ |
| MySQL Repeatable Read | ✅ | ✅ | ✅ | ❌ |
| MySQL Serializable | ✅ | ✅ | ✅ | ✅ |
| PostgreSQL Read Committed | ✅ | ❌ | ❌ | ❌ |
| PostgreSQL Repeatable Read | ✅ | ✅ | ✅ | ❌ |
| PostgreSQL Serializable | ✅ | ✅ | ✅ | ✅ |

### 5. 关键差异总结
* **隔离级别实现**：PostgreSQL使用多版本并发控制（MVCC）和依赖检测，而MySQL主要依赖锁机制。
* **Read Uncommitted**：在PostgreSQL中实际上等同于Read Committed，而MySQL则完全按照标准实现。
* **并发控制**：PostgreSQL在可重复读级别下对并发更新的控制更为严格，会直接拒绝可能导致数据不一致的操作，而MySQL则相对宽松。

### 6. 实践建议
* 在开发前详细了解所使用数据库的隔离级别实现细节。
* 使用高隔离级别时，务必实现适当的重试机制，以处理可能的并发冲突。
* 需要考虑并妥善处理事务超时和死锁情况。
* 在开发阶段进行充分的并发测试，特别是在使用较高隔离级别时。
* 在选择隔离级别时，需要在数据一致性要求和系统性能之间找到合适的平衡点。


## lecture 10 使用 github actions设置CI/CD(持续集成/持续部署)
1. 工作流配置
- 在`.github/workflows`目录下创建一个`ci.yml`文件
- 在`ci.yml`文件中配置工作流(推送到主分支或者pr时触发)