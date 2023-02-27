##分布式锁，实质就是redis中的键值对
实现方法：
1、先构造结构体，包含trylock、lock等方法，
在lock方法中，根据key生成对应的uuid作为val，必然会拿到lock
lock时会使用lua脚本get指定的key，有三种状态
    1、没有这个ket，那么直接set一个key
    2、key存在，同时是我的key
    3、key存在，不是我的key

在trylock方法中，使用setNX尝试拿到key/val，不一定成功
实际上，只要设置key/val成功，就能拿到lock
lock包含：key、val、超时时间等信息