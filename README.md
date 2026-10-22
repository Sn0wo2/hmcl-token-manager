# hmcl-token-manager

> 帮助你导入 / 导出 [HMCL](https://github.com/HMCL-dev/HMCL) Minecraft 账户`refresh_token`和`access_token`的小工具

> [!TIP]
> 如果你需要 Account Share 给朋友但是不想登录 Microsoft 账户的话应该会对你有帮助

### 引言

HMCL 在 [v3.16.0.347](https://docs.hmcl.net/changelog/dev.html#HMCL-3.16.0.347) 引入了 hmcl-obfuscated-v1 用于混淆`access_token`和`refresh_token`的存储, 并一起更新了一堆数据模型
在此之前我们只需要直接去 `%APPDATA%/.hmcl/` 找到 `accounts.json` 就可以导出账户信息. 现在账户数据经过混淆且结构有所变化, 手动迁移会比较麻烦, 因此有了这个工具

### 使用

启动程序后选择导入或导出账户. 导出时需要设置加密密码, 加密后的账户数据会复制到剪贴板; 导入时粘贴加密数据并输入密码即可. 如果剪贴板不可用, 导出的数据会打印到标准输出

默认情况下, 程序会根据操作系统查找 HMCL 用户目录, 也可以通过 `HMCL_USER_HOME` 环境变量指定目录
