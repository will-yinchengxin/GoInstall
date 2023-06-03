## Go os/exec 详解及问题
exec包执行外部命令。它包装了os.StartProcess函数以便更容易的修正输入和输出，使用管道连接I/O，以及作其它的一些调整。

```go
type Cmd struct {
    // Path是将要执行的命令的路径。
    //
    // 该字段不能为空，如为相对路径会相对于Dir字段。
    Path string
    // Args保管命令的参数，包括命令名作为第一个参数；如果为空切片或者nil，相当于无参数命令。
    //
    // 典型用法下，Path和Args都应被Command函数设定。
    Args []string
    // Env指定进程的环境，如为nil，则是在当前进程的环境下执行。
    Env []string
    // Dir指定命令的工作目录。如为空字符串，会在调用者的进程当前目录下执行。
    Dir string
    // Stdin指定进程的标准输入，如为nil，进程会从空设备读取（os.DevNull）
    Stdin io.Reader
    // Stdout和Stderr指定进程的标准输出和标准错误输出。
    //
    // 如果任一个为nil，Run方法会将对应的文件描述符关联到空设备（os.DevNull）
    //
    // 如果两个字段相同，同一时间最多有一个线程可以写入。
    Stdout io.Writer
    Stderr io.Writer
    // ExtraFiles指定额外被新进程继承的已打开文件流，不包括标准输入、标准输出、标准错误输出。
    // 如果本字段非nil，entry i会变成文件描述符3+i。
    //
    // BUG: 在OS X 10.6系统中，子进程可能会继承不期望的文件描述符。
    // http://golang.org/issue/2603
    ExtraFiles []*os.File
    // SysProcAttr保管可选的、各操作系统特定的sys执行属性。
    // Run方法会将它作为os.ProcAttr的Sys字段传递给os.StartProcess函数。
    SysProcAttr *syscall.SysProcAttr
    // Process是底层的，只执行一次的进程。
    Process *os.Process
    // ProcessState包含一个已经存在的进程的信息，只有在调用Wait或Run后才可用。
    ProcessState *os.ProcessState
    // 内含隐藏或非导出字段
}

// 函数返回一个*Cmd，用于使用给出的参数执行name指定的程序。返回值只设定了Path和Args两个参数。
// 如果name不含路径分隔符，将使用LookPath获取完整路径；否则直接使用name。参数arg不应包含命令名。
func Command(name string, arg ...string) *Cmd


// Run执行c包含的命令，并阻塞直到完成。
// 如果命令成功执行，stdin、stdout、stderr的转交没有问题，并且返回状态码为0，方法的返回值为nil；如果命令没有执行或者执行失败，会返回*ExitError类型的错误；否则返回的error可能是表示I/O问题。
// Run = Start + Wait
func (c *Cmd) Run() error


// Start开始执行c包含的命令，但并不会等待该命令完成即返回。Wait方法会返回命令的返回状态码并在命令返回后释放相关的资源
func (c *Cmd) Start() error

// Wait会阻塞直到该命令执行完成，该命令必须是被Start方法开始执行的。
// 如果命令成功执行，stdin、stdout、stderr的转交没有问题，并且返回状态码为0，方法的返回值为nil；
// 如果命令没有执行或者执行失败，会返回*ExitError类型的错误；否则返回的error可能是表示I/O问题。Wait方法会在命令返回后释放相关的资源。
func (c *Cmd) Wait() error

// 执行命令并返回标准输出的切片。
func (c *Cmd) Output() ([]byte, error)
````
在使用 Golang 的 exec.Exec() 函数执行命令时，如果你需要执行一个包含空格或特殊符号的命令字符串，那么你需要使用 "sh" "-c" 或者 "bash" "-c" 来解析命令字符串。

这是因为 exec.Exec() 函数默认会将命令字符串按照空格进行分割，然后将分割后的字符串数组作为参数传递给命令执行函数，这种方式不太适用于包含空格或特殊符号的命令字符串。

使用 "sh" "-c" 或者 "bash" "-c" 可以将整个命令字符串作为一个参数传递给 shell 执行，这样就可以正确解析包含空格或特殊符号的命令字符串了。

例如，如果你需要执行一个命令字符串 ls -l /tmp，你可以使用以下代码：

```go
cmd := exec.Command("sh", "-c", "ls -l /tmp")
output, err := cmd.Output()
````

这样就可以正确执行包含空格的命令字符串了。