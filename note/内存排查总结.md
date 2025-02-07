参考: https://www.51cto.com/article/612789.html

- 首先使用 pmap 查看其中任意一个 worker 进程的内存分布，这里是 4199，使用 pmap 命令的输出如下所示。

```
pmap -x <pid> | sort -k 3 -n -r  
00007f2340539000 475240 461696 461696 rw--- [ anon ]  
```

- 随后使用 cat /proc/<pid>/smaps | grep 7f2340539000 查找某一段内存的起始和结束地址，如下所示。

```
cat /proc/<pid>/smaps | grep 7f2340539000  
7f2340539000-7f235d553000 rw-p 00000000 00:00 0 
```

- 随后使用 gdb 连上这个进程，dump 出这一段内存。


```
gdb -pid 4199  
dump memory memory.dump 0x7f2340539000 0x7f235d553000 
```

- 随后使用 strings 命令查看这个 dump 文件的可读字符串内容
```
strings memory.dump
```