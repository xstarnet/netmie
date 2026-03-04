# 移动端 NetBird 和 V2Ray 互斥连接 API

## 概述

移动端（Android 和 iOS）现在支持 NetBird P2P VPN 和 V2Ray 代理的互斥连接模式。由于移动系统限制，同一时间只能有一个活跃的 VPN 连接。

## 核心 API

### 1. ConnectNetBird()

启动 NetBird 连接。如果 V2Ray 正在运行,会自动断开。

**Android 示例:**
```kotlin
val client = Client(...)
try {
    client.connectNetBird()
    println("NetBird connected")
} catch (e: Exception) {
    println("Failed to connect: ${e.message}")
}
```

**iOS 示例:**
```swift
let client = NetBirdSDKClient(...)
do {
    try client.connectNetBird()
    print("NetBird connected")
} catch {
    print("Failed to connect: \(error)")
}
```

### 2. ConnectV2Ray(configPath: String)

启动 V2Ray 连接。如果 NetBird 正在运行,会自动断开。

**参数:**
- `configPath`: V2Ray 配置文件的完整路径

**Android 示例:**
```kotlin
val configPath = "${context.filesDir}/v2ray-config.json"
try {
    client.connectV2Ray(configPath)
    println("V2Ray connected")
} catch (e: Exception) {
    println("Failed to connect: ${e.message}")
}
```

**iOS 示例:**
```swift
let configPath = "\(documentsDirectory)/v2ray-config.json"
do {
    try client.connectV2Ray(configPath)
    print("V2Ray connected")
} catch {
    print("Failed to connect: \(error)")
}
```

### 3. Disconnect()

断开当前活跃的连接（NetBird 或 V2Ray）。

**Android 示例:**
```kotlin
try {
    client.disconnect()
    println("Disconnected")
} catch (e: Exception) {
    println("Failed to disconnect: ${e.message}")
}
```

**iOS 示例:**
```swift
do {
    try client.disconnect()
    print("Disconnected")
} catch {
    print("Failed to disconnect: \(error)")
}
```

### 4. GetConnectionStatus()

获取当前连接状态（JSON 字符串）。

**返回格式:**
```json
{
  "type": "NetBird",
  "connected": true,
  "details": {
    "peers": 5,
    "fqdn": "device.netbird.cloud",
    "ip": "100.64.0.1"
  }
}
```

或者（V2Ray）:
```json
{
  "type": "V2Ray",
  "connected": true,
  "details": {
    "status": "running",
    "config_version": "1.0"
  }
}
```

**Android 示例:**
```kotlin
val statusJson = client.getConnectionStatus()
val status = JSONObject(statusJson)
println("Type: ${status.getString("type")}")
println("Connected: ${status.getBoolean("connected")}")
```

**iOS 示例:**
```swift
let statusJson = client.getConnectionStatus()
if let data = statusJson.data(using: .utf8),
   let status = try? JSONSerialization.jsonObject(with: data) as? [String: Any] {
    print("Type: \(status["type"] ?? "unknown")")
    print("Connected: \(status["connected"] ?? false)")
}
```

## 配置管理 API

### 5. SetV2RayConfig(configJSON: String)

设置 V2Ray 配置（JSON 字符串）。

**Android 示例:**
```kotlin
val config = """
{
  "inbounds": [
    {
      "port": 10808,
      "protocol": "socks",
      "settings": {
        "auth": "noauth",
        "udp": true
      }
    }
  ],
  "outbounds": [
    {
      "protocol": "vmess",
      "settings": {
        "vnext": [...]
      }
    }
  ]
}
"""
try {
    client.setV2RayConfig(config)
    println("Config saved")
} catch (e: Exception) {
    println("Failed to save config: ${e.message}")
}
```

### 6. GetV2RayConfig()

获取当前 V2Ray 配置（JSON 字符串）。

**Android 示例:**
```kotlin
try {
    val config = client.getV2RayConfig()
    println("Config: $config")
} catch (e: Exception) {
    println("Failed to get config: ${e.message}")
}
```

## 使用流程

### 典型场景 1: 从 NetBird 切换到 V2Ray

```kotlin
// 1. 用户点击 "连接 V2Ray"
val configPath = "${context.filesDir}/v2ray-config.json"

// 2. 调用 ConnectV2Ray (会自动断开 NetBird)
client.connectV2Ray(configPath)

// 3. 检查状态
val status = JSONObject(client.getConnectionStatus())
if (status.getString("type") == "V2Ray" && status.getBoolean("connected")) {
    println("Successfully switched to V2Ray")
}
```

### 典型场景 2: 从 V2Ray 切换到 NetBird

```kotlin
// 1. 用户点击 "连接 NetBird"
// 2. 调用 ConnectNetBird (会自动断开 V2Ray)
client.connectNetBird()

// 3. 检查状态
val status = JSONObject(client.getConnectionStatus())
if (status.getString("type") == "NetBird" && status.getBoolean("connected")) {
    println("Successfully switched to NetBird")
}
```

### 典型场景 3: 配置并连接 V2Ray

```kotlin
// 1. 设置配置
val config = loadV2RayConfigFromServer()
client.setV2RayConfig(config)

// 2. 连接
val configPath = "${context.filesDir}/v2ray-config.json"
client.connectV2Ray(configPath)

// 3. 验证
val status = JSONObject(client.getConnectionStatus())
println("Connected: ${status.getBoolean("connected")}")
```

## V2Ray 配置注意事项

### 移动端使用 Proxy 模式

移动端默认使用 **Proxy 模式**（SOCKS5/HTTP），而不是 TUN 模式，以避免与系统 VPN 冲突。

**推荐配置:**
```json
{
  "inbounds": [
    {
      "port": 10808,
      "protocol": "socks",
      "settings": {
        "auth": "noauth",
        "udp": true
      }
    },
    {
      "port": 10809,
      "protocol": "http",
      "settings": {}
    }
  ],
  "outbounds": [...]
}
```

### 配置文件路径

- **Android**: `${context.filesDir}/v2ray-config.json`
- **iOS**: `${documentsDirectory}/v2ray-config.json`

### 端口选择

- SOCKS5: 推荐 10808
- HTTP: 推荐 10809
- 避免使用系统保留端口（< 1024）

## 错误处理

所有 API 都会抛出异常,需要适当处理:

```kotlin
try {
    client.connectV2Ray(configPath)
} catch (e: Exception) {
    when {
        e.message?.contains("config file not found") == true -> {
            // 配置文件不存在
            showError("请先配置 V2Ray")
        }
        e.message?.contains("port") == true -> {
            // 端口被占用
            showError("端口被占用,请检查配置")
        }
        else -> {
            // 其他错误
            showError("连接失败: ${e.message}")
        }
    }
}
```

## 线程安全

所有 API 都是线程安全的,可以从任何线程调用。内部使用互斥锁保护连接状态。

## 性能考虑

- `GetConnectionStatus()` 是轻量级操作,可以频繁调用
- 连接切换操作（ConnectNetBird/ConnectV2Ray）会阻塞直到完成
- 建议在后台线程执行连接操作

## 示例应用集成

### Android Activity 示例

```kotlin
class MainActivity : AppCompatActivity() {
    private lateinit var client: Client

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        client = Client(...)

        btnNetBird.setOnClickListener {
            lifecycleScope.launch(Dispatchers.IO) {
                try {
                    client.connectNetBird()
                    withContext(Dispatchers.Main) {
                        updateUI("NetBird Connected")
                    }
                } catch (e: Exception) {
                    withContext(Dispatchers.Main) {
                        showError(e.message)
                    }
                }
            }
        }

        btnV2Ray.setOnClickListener {
            lifecycleScope.launch(Dispatchers.IO) {
                try {
                    val configPath = "${filesDir}/v2ray-config.json"
                    client.connectV2Ray(configPath)
                    withContext(Dispatchers.Main) {
                        updateUI("V2Ray Connected")
                    }
                } catch (e: Exception) {
                    withContext(Dispatchers.Main) {
                        showError(e.message)
                    }
                }
            }
        }
    }
}
```

### iOS ViewController 示例

```swift
class ViewController: UIViewController {
    var client: NetBirdSDKClient!

    override func viewDidLoad() {
        super.viewDidLoad()

        client = NetBirdSDKClient(...)
    }

    @IBAction func connectNetBird(_ sender: Any) {
        DispatchQueue.global().async {
            do {
                try self.client.connectNetBird()
                DispatchQueue.main.async {
                    self.updateUI("NetBird Connected")
                }
            } catch {
                DispatchQueue.main.async {
                    self.showError(error.localizedDescription)
                }
            }
        }
    }

    @IBAction func connectV2Ray(_ sender: Any) {
        DispatchQueue.global().async {
            do {
                let configPath = "\(self.documentsDirectory)/v2ray-config.json"
                try self.client.connectV2Ray(configPath)
                DispatchQueue.main.async {
                    self.updateUI("V2Ray Connected")
                }
            } catch {
                DispatchQueue.main.async {
                    self.showError(error.localizedDescription)
                }
            }
        }
    }
}
```

## 调试

启用详细日志:

```kotlin
// Android
client.setTraceLogLevel()

// iOS
client.setTraceLogLevel()
```

查看日志输出:
- Android: `adb logcat | grep V2Ray`
- iOS: Xcode Console

## 限制

1. 同一时间只能有一个连接活跃
2. V2Ray 使用 Proxy 模式,需要应用层配置代理
3. 配置文件必须是有效的 JSON 格式
4. 端口必须未被占用

## 故障排除

### 问题: "config file not found"
**解决**: 先调用 `SetV2RayConfig()` 设置配置

### 问题: "port already in use"
**解决**: 修改配置文件中的端口号

### 问题: 切换连接失败
**解决**: 先调用 `Disconnect()` 再连接新的

### 问题: GetConnectionStatus 返回 "None"
**解决**: 检查是否成功调用了 Connect 方法
