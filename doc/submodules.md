# git依赖关系工具

**解决公共代码问题**。如果某些文件，在项目A和项目B中都会用到，例如组件库，那么这些文件可以使用公共代码管理工具来管理，减少重复代码。

##### 1.git submodule和git subtree有什么区别？

##### 1.1 仓库独立性：

git submodule：子模块保持独立的仓库状态，拥有自己的提交历史和分支。这意味着子模块可以独立于主项目进行开发、维护和版本控制。
git subtree：子树将子仓库的内容合并到主项目的仓库中，不保留独立的仓库。因此，子树的内容与主项目共享同一个提交历史和分支。

##### 1.2 初始化和更新：

git submodule：使用子模块需要执行额外的初始化和更新命令。在克隆包含子模块的项目时，需要特别注意子模块的初始化和更新操作。
git subtree：使用子树不需要额外的初始化和更新命令。子树的内容作为主项目的一部分被直接管理和更新。

##### 1.3 仓库结构：

git submodule：主项目和子模块的仓库分别存在，可能会导致仓库冗余和复杂性增加。然而，这种结构也保留了子模块的独立性，便于单独管理和维护。
git subtree：合并子仓库的内容后，主项目仓库不会出现子仓库的文件夹，仓库结构更加整洁。但这也意味着子仓库的内容不再独立，难以单独分离出来。

##### 1.4 适用场景：

git submodule：适用于需要独立开发和维护子模块的场景，例如当主项目依赖于其他外部仓库或库时，或者当多个项目需要共享一些通用的代码库时。
git subtree：适用于需要将外部仓库的特定部分集成到主项目中，并且不需要独立开发和维护子仓库的场景。例如，当主项目和子项目之间共享部分代码时，可以使用子树来管理这个集成过程。

##### 2.了解 git submodules

有2个概念：**主项目、submodule（子模块）**。这两个都是各自完整的git仓库。

##### 2.1 如何让一个git仓库变为另一个git仓库的submodule

> git submodule add ...(仓库B的地址，即git clone时后面那串东西)

执行操作后，会在当前父项目下新建个文件夹，名字就是 submodule 仓库的名字。这个文件夹里面的内容，是 submodule 对应 Git 仓库的完整代码。如果你希望换个名字，或者换个路径（例如放在某个更深的目录下），也是允许的，需要后面增加个路径参数，例如git submodule add ...(仓库地址) src/B(你希望 submodule 位于的文件夹路径)

##### 2.2 submodule 的父子关系存在哪里

关系保存在主项目的git仓库中的.gitmodules文件中。这个文件中主要记录了子模块的url，如果添加的时候使用ssh链接，那url就是ssh,如果是http链接，这个url就是https。

> ```go
> [submodule "git-learn-submodules1"]
>     path = git-learn-submodules1
>     url = https://github.com/js-xzw7/git-learn-submodules1.git
> [submodule "git-learn-submodules2"]
>     path = git-learn-submodules2
>     url = https://github.com/js-xzw7/git-learn-submodules2.git
> ```

在git上的目录，其中两个模块是用hash commit 来维护,指向的是两个子仓库的地址

> | [git-learn-submodules1 @ a67f56f](https://github.com/js-xzw7/git-learn-submodules1/tree/a67f56f1a899557bd7a6beebb11be8341580c862) | [add submodules](https://github.com/js-xzw7/git-learn/commit/f62b4cf6608c24b30ea00b73890017bdaf3b67b1) | 1 minute ago  |
> | ------------------------------------------------------------ | ------------------------------------------------------------ | ------------- |
> | [git-learn-submodules2 @ f99fdf7](https://github.com/js-xzw7/git-learn-submodules2/tree/f99fdf79a50ec1b9ac291d40f42953f6e67caefb) | [add submodules](https://github.com/js-xzw7/git-learn/commit/f62b4cf6608c24b30ea00b73890017bdaf3b67b1) | 1 minute ago  |
> | [.gitmodules](https://github.com/js-xzw7/git-learn/blob/main/.gitmodules) | [add submodules](https://github.com/js-xzw7/git-learn/commit/f62b4cf6608c24b30ea00b73890017bdaf3b67b1) | 1 minute ago  |
> | [README.md](https://github.com/js-xzw7/git-learn/blob/main/README.md) | [Create README.md](https://github.com/js-xzw7/git-learn/commit/8b3ef9c735bb52c9bb6beb2961d5b7f7607c3022) | 5 minutes ago |

##### 3.实际开发操作

##### 3.1 初始化和拉取子模块

1.**克隆包含子模块的主仓库**：

> git clone --recurse-submodules <repository-url>
>
> --recurse-submodules 初始化子模块，若使用，克隆后需要初始化一下子模块

2.**初始化子模块**：

> git submodule init

3.**拉取子模块的代码**：

> git submodule update



##### 3.2 更新子模块到最新提交

1. **进入子模块目录**：

   ```
   cd <path_to_submodule>
   ```

2. **获取最新的提交**：

   ```
   git fetch
   ```

3. **切换到你想要的分支（通常是 `master` 或 `main`）**：

   ```
   git checkout master
   ```

4. **拉取最新的提交**：

   ```
   git pull origin master
   ```

5. **返回主项目目录**：

   ```
   cd ..
   ```

6. **更新主项目中的子模块引用**：

   ```go
   git add <path_to_submodule>
   git commit -m "Update submodule to latest version"
   ```

##### 3.2 更新所有子模块

如果你的项目中有多个子模块，可以使用以下命令一次性更新所有子模块：

```go
git submodule update --remote
```



#### subTree

##### 1.引用子项目

>  git subtree add --prefix=proto/dream_dance_tiktok   *<path_to_subtree>*  main --squash
>
> `--squash` 参数表示不拉取历史信息，而只生成一条 commit 信息，这是一个可选参数，可以不加。

##### 2.更新子项目并提交

在开发过程中修改了子项目，可以在主项目中提交子项目的改变到它自己的仓库中(先将主项目中的变化提交到远程仓库，在执行如下操作)

> git subtree push --prefix=proto/dream_dance_tiktok  *<path_to_subtree>* main

##### 3.子项目发生了变化，其他微服务通过如下指令可以更新代码

> git subtree pull --prefix=proto/dream_dance_tiktok  *<path_to_subtree>* main --squash



**这三个指令基本上就能应付日常的大部分操作了，不过每次都要输入一个长长的地址很不方便，我们可以给地址取一个别名：**

> git remote add -f protoServer  *<path_to_subtree>*

简化命令

> git subtree add --prefix=proto/dream_dance_tiktok protoServer  main --squash 
>
> git subtree pull --prefix=proto/dream_dance_tiktok protoServer  main --squash
>
>  git subtree push --prefix=proto/dream_dance_tiktok protoServer  main 
