# 线程

- 创建线程的方法

1. 继承Thread，重写run
2. new一个Thread的时候传入Runnable
3. 向ExecutorService提交Callable
4. CompletableFuture
....
归根到底都是通过Thread的start0方法创建线程

- 线程状态

1. new
2. runnable: io
3. running
4. blocked
5. wait
6. time_wait
7. terminated

- 上下文切换

将线程执行的上下文，比如 寄存器中的变量，pc指针，栈信息等存在内存中，当线程切换的时候，需要将内存中的上下文切换到寄存器中，这个过程称为上下文切换

- 并发和并行
TODO

- 同步和异步
TODO

- 线程池

1. 核心线程数
2. 总线程数
3. 非核心线程存活时间
4. 任务队列
5. 线程工厂
6. 拒绝策略

- 线程池等拒绝策略
1. 抛出异常 (默认)
2. 谁提交谁执行
3. 抛弃当前任务
4. 丢弃最旧的任务
5. 自定义


- 死锁

死锁的条件：
1. 资源互斥：一个资源只能被有限的线程访问                     (减少锁的争夺，比如 采用无锁算法)
2. 占有且等待：占有了部分锁资源后，等待其他锁时不会释放。       (一开始获得所有锁资源，失败就释放)
3. 不可剥夺：锁不能被剥夺，只能主动释放                      (强制介入，释放锁)
4. 循环等待：等待                                        (为资源设置有序的编号，按顺序获取)





- synchronized

