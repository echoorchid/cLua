# cLua
lua的代码覆盖率工具

# 特性
* C++开发，性能更高
* 简单require即可使用，或通过[hookso](https://github.com/esrrhs/hookso)注入
* 独立的命令行解析工具

# 编译
* 编译libclua.so
```
# cmake .
# make
```
* 编译clua解析工具
```
# go build clua.go
```

# 使用
* lua文件里使用如下
```
-- 加载libclua.so
local cl = require "libclua"
-- 开始记录执行过程，生成结果文件
-- 第一个参数为结果文件的文件名
-- 第二个参数为定时生成结果文件的间隔（秒），0表示关闭
cl.start("test.cov", 5)

-- 执行某些事情
do_something()

-- 结束记录
cl.stop()
```
* 然后用clua解析结果，clua更多参数参考-h
```
# ./clua -i test.cov
```

# 示例
* 运行test.lua
```
# lua5.3 ./test.lua
```
* 查看目录下，已有test.cov文件
```
# ll test.cov
```
* 查看结果，每行前面的数字表示执行的次数，空表示没被执行
```
# ./clua -i test.cov     
total points = 20, files = 1
coverage of /home/project/clua/test.lua:
    local cl = require "libclua"
    cl.start("test.cov", 5)
    
1   function test1(i)
10      if i % 2 then
10          print("a "..i)
        else
            print("b "..i)
        end
11  end
    
1   function test2(i)
40      if i > 30 then
19          print("c "..i)
        else
21          print("d "..i)
        end
41  end
    
1   function test3(i)
    
51      if i > 0 then
51          print("e "..i)
        else
            print("f "..i)
        end
    
52  end
    
102 for i = 0, 100 do
101     if i < 10 then
10          test1(i)
91      elseif i < 50 then
40          test2(i)
        else
51          test3(i)
        end
    end
    
1   cl.stop()

/home/project/clua/test.lua total coverage 60%
```
* 最后一行输出文件的总体覆盖率，这个因为有else、end之类的影响，所以并不完全准确
