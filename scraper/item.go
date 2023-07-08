package scraper

type Category struct {
	Identity string
	Name     string
}

type TagItem struct {
	Identity string
	Name     string
}

type Tag struct {
	Category Category
	Item     []TagItem
}

type Character struct {
	Name         string
	Introduction string
	Avatar       string
	Images       []string
}

type Item struct {
	proxy       string
	Name        string      // 名称
	Cover       string      // 封面
	Preview     []string    // 预览图
	Tags        []Tag       // 标签
	Brand       string      // 品牌
	ReleaseDate string      // 发售日
	Link        string      // 官网
	Information []string    // 介绍页面
	SaveData    string      // 存档
	WalkThrough string      // 攻略
	Size        string      // 大小（仅供参考）
	Magnet      string      // 磁力链接
	BtFile      string      // bt 种子
	OtherInfo   string      // 其它信息
	Origin      string      // 来源网站
	Character   []Character // 角色
	Genre       []string    // 类别
	Story       string      // 故事简介
}
