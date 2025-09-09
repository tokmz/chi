package chi

type Config struct {
}

type ServerConfig struct {
	// 服务器监听地址
	Addr string `json:"addr" yaml:"addr"`
	// 服务器模式
	Mode string `json:"mode" yaml:"mode"`
	// 服务器名称
	Name string `json:"name" yaml:"name"`
	// 服务器描述
	Desc string `json:"desc" yaml:"desc"`
	// 默认上传文件目录
	Upload string `json:"upload" yaml:"upload"`
	// 服务器版本
	Version string `json:"version" yaml:"version"`
}
