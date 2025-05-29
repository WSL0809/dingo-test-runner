# MySQL Tester

This is a golang implementation of [MySQL Test Framework](https://github.com/mysql/mysql-server/tree/8.0/mysql-test).

## Requirements

- All the tests should be put in [`t`](./t), take [t/example.test](./t/example.test) as an example.
- All the expected test results should be put in [`r`](./r). Result file has the same file name with the corresponding test file, but with a default `.result` file extension, it can be changed by `-extension`, take [r/example.result](./r/example.result) as an examle.

## How to use

Build the `mysql-tester` binary:
```sh
make
```

Basic usage:
```
Usage of ./mysql-tester:
  -all
        run all tests
  -host string
        The host of the TiDB/MySQL server. (default "127.0.0.1")
  -log-level string
        The log level of mysql-tester: info, warn, error, debug. (default "error")
  -params string
        Additional params pass as DSN(e.g. session variable)
  -passwd string
        The password for the user.
  -port string
        The listen port of TiDB/MySQL server. (default "4000")
  -record
        Whether to record the test output to the result file.
  -reserve-schema
        Reserve schema after each test
  -retry-connection-count int
        The max number to retry to connect to the database. (default 120)
  -user string
        The user for connecting to the database. (default "root")
  -xunitfile string
        The xml file path to record testing results.
  -check-error
        If --error ERR does not match, return error instead of just warn
  -extension
        Specify the extension of result file under special requirement, default as ".result"
```

By default, it connects to the TiDB/MySQL server at `127.0.0.1:4000` with `root` and no passward:
```sh
./mysql-tester # run all the tests
./mysql-tester example # run a specified test
./mysql-tester example1 example2   example3 # seperate different tests with one or more spaces
# modify current example cases for .result output.
./mysql-tester -record=1 -check-error=1
./mysql-tester -record=1 -host=127.0.0.1 -port=3306 -user=root -passwd=123456
```

For more details about how to run and write test cases, see the [Wiki](https://github.com/pingcap/mysql-tester/wiki) page.

## 生成测试报告

使用以下命令可以生成 JUnit XML 格式的测试报告：

```sh
./mysql-tester --report --junit-results-dir=./junit-results [测试文件]
```

allure generate junit-results -o ./allure-output --clean

allure open allure-output

 ./mysql-tester -record=1 -host=172.30.14.172 -port=3307 -user=root -passwd=123123 quickbi/interval

## 版本历史

### v1.0.2 (2025-05-29)

- **数据库连接管理优化**：
  - 创建了专门的连接管理模块 `ConnectionManager`，负责管理所有数据库连接
  - 使用连接池技术优化了数据库连接性能
  - 重构了连接相关的代码，提高了代码的模块化程度和可维护性
  - 优化了连接资源的使用，减少了不必要的连接创建和销毁