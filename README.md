# rsdish

为移动硬盘设计的rclone脚本自动生成工具。跨平台。动态扫描硬盘位置。rsdish本身并不会删除任何你的文件，它只是一个脚本生成工具。

_注意：使用前建议将rclone和rsdish添加到PATH_

## 例子

### 创建库

1. 创建特定文件夹。在E盘下创建"E:/volumes/movie_volume";
2. 进入"E:/volumes/movie_volume"，运行`rsdish template new`，这会在在当前文件夹创建一个volume.toml文件，包含一个随机生成的uuid作为当前volume所属library的标识;
3. 如果你要为已经存在的library附加volume，那么可以运行`rsdish template new --from <UUID>`（或者`rsdish template new --from <SHORT>`，详情见“收藏library”）;

### 扫描

1. 要查看当前系统所有的library和它们从属的volume的信息，运行`rsdish scan lib`;

### 生成同步脚本

_注意：rsdish中的sync概念更类似于rclone中的copy的概念，同属于一个library的volume会相互发送彼此没有的文件，而不是删除对方没有的文件。如果你想要删除某个文件，详情见“删除文件”_

1. 运行`rsdish sync`或者`rsdish sync <>`