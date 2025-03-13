# mydocker
Go~It's my Docker!!!

学习用Go实现一个简易Docker，[代码参考](https://github.com/lixd/mydocker?tab=readme-ov-file).

### TODO
- 添加命令 `start`
    - 缺少网络配置
- `docker ps` 命令增加参数`-a` ，之后的`docker ps` 只会列出RUNNING的容器，`docker ps -a ` 会列出所有的容器
- 设置docker image的默认启动命令，启动命令应该是存储在了镜像中的配置文件`config.json`中
- Cgroups 的控制逻辑不够完善：
    1. 只有在交互式`-it`的情况下才会进行资源控制
    2. 只有一个 docker hierarchy
    3. `exec` 无法进行资源控制