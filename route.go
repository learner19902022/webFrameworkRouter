package web

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

type router struct {
	// trees 是按照 HTTP 方法来组织的
	// 如 GET => *node
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// addRoute 注册路由。
// method 是 HTTP 方法
// - 已经注册了的路由，无法被覆盖。例如 /user/home 注册两次，会冲突[already given]
// - path 必须以 / 开始并且结尾不能有 /，中间也不允许有连续的 / [already given]
// - 不能在同一个位置注册不同的参数路由，例如 /user/:id 和 /user/:name 冲突 [already given]
// - 不能在同一个位置同时注册通配符路由和参数路由，例如 /user/:id 和 /user/* 冲突 [already given]
// - 同名路径参数，在路由匹配的时候，值会被覆盖。例如 /user/:id/abc/:id，那么 /user/123/abc/456 最终 id = 456
func (r *router) addRoute(method string, path string, handler HandleFunc) {
	//避免空路由
	if path == "" {
		panic("web: 路由是空字符串")
	}
	//保证路由地址由/开头
	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}
	//避免路由地址由/结尾
	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}

	root, ok := r.trees[method]
	// 这是一个全新的 HTTP 方法，创建根节点
	if !ok {
		// 创建根节点
		root = &node{path: "/"}
		r.trees[method] = root
	}
	//判断是否为根结点路由，如果未注册句柄过则注册句柄
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突[/]")
		}
		root.handler = handler
		return
	}

	//去除第一个/，并且切分
	segs := strings.Split(path[1:], "/") //有空格行么？有特殊字符行么
	// 开始一段段处理
	for _, s := range segs {
		if s == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		root = root.childOrCreate(s)
	}
	//如果已经注册过句柄，则需要panic
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	//如果错误注册了路由，应该使用什么接口修改？无法重新写入吧？
	root.handler = handler
}

// findRoute 查找对应的节点
// 注意，返回的 node 内部 HandleFunc 不为 nil 才算是注册了路由 //难道不是"才算是找到了路由"？
func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return &matchInfo{n: root}, true //root.handler != nil
	}

	segs := strings.Split(strings.Trim(path, "/"), "/")
	mi := &matchInfo{}
	for _, s := range segs {
		var matchParam, matchRegex bool
		root, matchParam, matchRegex, ok = root.childOf(s)
		if !ok {
			return nil, false
		}
		//命中参数路由
		if matchParam {
			mi.addValue(root.path[1:], s)
		}
		if matchRegex {
			mi.addValue(root.path[1:strings.Index(root.path, "(")], s)
		}
		if !(matchParam || matchRegex) && root.typ == nodeTypeAny && root.children == nil && root.regChild == nil && root.paramChild == nil && root.starChild == nil {
			mi.n = root
			return mi, true
		}
	}
	mi.n = root
	return mi, true //root.handler != nil
}

type nodeType int

const (
	// 静态路由
	nodeTypeStatic = iota
	// 正则路由
	nodeTypeReg
	// 路径参数路由
	nodeTypeParam
	// 通配符路由
	nodeTypeAny
)

// node 代表路由树的节点
// 路由树的匹配顺序是：
// 1. 静态完全匹配
// 2. 路径参数匹配：形式 :param_name
// 3. 通配符匹配：*
// 这是不回溯匹配
type node struct {
	typ nodeType

	path string
	// children 子节点
	// 子节点的 path => node
	children map[string]*node
	// handler 命中路由之后执行的逻辑
	handler HandleFunc

	// 通配符 * 表达的节点，任意匹配
	starChild *node

	paramChild *node
	// 正则路由和参数路由都会使用这个字段
	paramName string

	// 正则表达式
	regChild *node
	regExpr  *regexp.Regexp
}

// childOrCreate 查找子节点，
// 首先会判断 path 是不是通配符路径
// 其次判断 path 是不是参数路径，即以 : 开头的路径
// 最后会从 children 里面查找，
// 如果没有找到，那么会创建一个新的节点，并且保存在 node 里面
func (n *node) childOrCreate(path string) *node {
	if path == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}
		if n.regChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [%s]", path))
		}
		if n.starChild == nil {
			n.starChild = &node{
				path: path,
				typ:  nodeTypeAny,
			}
		}
		return n.starChild
	}

	if path[0] == ':' && path[len(path)-1] == ')' && strings.Contains(path, "(") {
		if n.starChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [%s]", path))
		}
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [%s]", path))
		}
		if n.regChild != nil {
			if n.regChild.path != path {
				panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.path, path))
			}
		} else {
			markIndex := strings.Index(path, "(")
			if string(path[markIndex+1]) == ")" {
				panic(fmt.Sprintf("web: 正则路由的正则规则不能为空"))
			}
			n.regChild = &node{
				path:      path,
				typ:       nodeTypeReg,
				paramName: path[1:markIndex],
				regExpr:   regexp.MustCompile(path[markIndex+1 : len(path)-1]),
			}
		}
		return n.regChild
	}

	// 以 : 开头，我们认为是参数路由
	if path[0] == ':' {
		if n.starChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}
		if n.regChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [%s]", path))
		}
		if n.paramChild != nil {
			if n.paramChild.path != path {
				panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.paramChild.path, path))
			}
		} else {
			n.paramChild = &node{
				path:      path,
				typ:       nodeTypeParam,
				paramName: path[1:],
			}
		}
		return n.paramChild
	}

	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		child = &node{
			path: path,
			typ:  nodeTypeStatic,
		}
		n.children[path] = child
	}
	return child
}

// child 返回子节点
// 第一个返回值 *node 是命中的节点
// 第二个返回值 bool 代表是否命中参数路由
// 第三个返回值 bool 代表是否命中正则路由
// 第四个返回值 bool 代表是否命中
func (n *node) childOf(path string) (*node, bool, bool, bool) {
	if path == "" { //需要对空字段支持正则路由命中么？
		return nil, false, false, false
	}
	if n.children == nil || func() bool {
		_, ok := n.children[path]
		return !ok
	}() {
		if n.regChild != nil {
			if n.regChild.regExpr.FindString(path) != "" { //len(n.regExpr.FindString(path)) == len(path) { //这边有点问题：如果正则规则里面有倾向于取更少的字符，那么会导致即使全匹配，也会有限选择更短的string；可是如果不是检查长度的话，又无法排除path中的一部分匹配到了正则规则
				return n.regChild, false, true, true
			}
			if n.paramChild != nil {
				return n.paramChild, true, false, true
			}
			return n.starChild, false, false, n.starChild != nil
		}
		if n.paramChild != nil {
			return n.paramChild, true, false, true
		}
		return n.starChild, false, false, n.starChild != nil
	}
	res, ok := n.children[path]
	return res, false, false, ok
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		// 大多数情况，参数路径只会有一段
		m.pathParams = map[string]string{key: value}
	}
	m.pathParams[key] = value
}

func (r *router) PrintAllRouters() { //DFS
	for method, root := range r.trees {
		fmt.Printf("======================\n")
		fmt.Printf("打印路由树，树名为 %s:\n", method)
		root.printNode("/")
	}
}

func (n *node) printNode(concatPath string) (upLayerPath string) {
	fmt.Printf("--------------\n")
	fmt.Printf("路由节点名为：%s\n", n.path)
	if n.path != "/" {
		concatPath = concatPath + "/" + n.path
	}
	switch n.typ {
	case nodeTypeStatic:
		fmt.Printf("路由节点类型为：%s\n", "静态路由节点")
		fmt.Printf("路由地址：%s\n", concatPath)
		fmt.Printf("句柄为 %s\n", runtime.FuncForPC(reflect.ValueOf(n.handler).Pointer()).Name())
	case nodeTypeReg:
		fmt.Printf("路由节点类型为：%s\n", "正则路由节点")
		fmt.Printf("路由地址：%s\n", concatPath)
		fmt.Printf("句柄为 %s\n", runtime.FuncForPC(reflect.ValueOf(n.handler).Pointer()).Name())
		fmt.Printf("正则表达式为：%s\n", n.regExpr.String())
	case nodeTypeParam:
		fmt.Printf("路由节点类型为：%s\n", "参数路由节点")
		fmt.Printf("路由地址：%s\n", concatPath)
		fmt.Printf("句柄为 %s\n", runtime.FuncForPC(reflect.ValueOf(n.handler).Pointer()).Name())
		fmt.Printf("参数名为：%s\n", n.paramName)
	case nodeTypeAny:
		fmt.Printf("路由节点类型为：%s\n", "通配符路由节点")
		fmt.Printf("路由地址：%s\n", concatPath)
		fmt.Printf("句柄为 %s\n", runtime.FuncForPC(reflect.ValueOf(n.handler).Pointer()).Name())
	}

	if concatPath == "/" {
		concatPath = ""
	}

	for _, childNode := range n.children {
		concatPath = childNode.printNode(concatPath)
	}

	if n.regChild != nil {
		concatPath = n.regChild.printNode(concatPath)
	}

	if n.paramChild != nil {
		concatPath = n.paramChild.printNode(concatPath)
	}

	if n.starChild != nil {
		concatPath = n.starChild.printNode(concatPath)
	}

	return strings.TrimSuffix(concatPath, "/"+n.path)
}

func (r *router) VerifyRouter(method string, testPath string, wantedRouteNode *node) (string, bool) {
	if wantedRouteNode == nil {
		panic("想测试的路由不能为空")
	}

	testNode, found := r.findRoute(method, testPath)
	if !found {
		return "未找到测试路径可使用的路由", false
	}

	msg, verified := testNode.n.equal(wantedRouteNode)

	return msg, verified
}

func (n *node) equal(y *node) (string, bool) {
	if y == nil {
		return "目标节点为 nil", false
	}
	if n.path != y.path {
		return fmt.Sprintf("%s 节点 path 不相等 x %s, y %s", n.path, n.path, y.path), false
	}

	nhv := reflect.ValueOf(n.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("%s 节点 handler 不相等 x %s, y %s", n.path, nhv.Type().String(), yhv.Type().String()), false
	}

	if n.typ != y.typ {
		return fmt.Sprintf("%s 节点类型不相等 x %d, y %d", n.path, n.typ, y.typ), false
	}

	if n.paramName != y.paramName {
		return fmt.Sprintf("%s 节点参数名字不相等 x %s, y %s", n.path, n.paramName, y.paramName), false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("%s 子节点长度不等", n.path), false
	}
	if len(n.children) == 0 {
		return "", true
	}

	if n.starChild != nil {
		str, ok := n.starChild.equal(y.starChild)
		if !ok {
			return fmt.Sprintf("%s 通配符节点不匹配 %s", n.path, str), false
		}
	}
	if n.paramChild != nil {
		str, ok := n.paramChild.equal(y.paramChild)
		if !ok {
			return fmt.Sprintf("%s 路径参数节点不匹配 %s", n.path, str), false
		}
	}

	if n.regChild != nil {
		str, ok := n.regChild.equal(y.regChild)
		if !ok {
			return fmt.Sprintf("%s 正则节点不匹配 %s", n.path, str), false
		}
	}

	for k, v := range n.children {
		yv, ok := y.children[k]
		if !ok {
			return fmt.Sprintf("%s 目标节点缺少子节点 %s", n.path, k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return n.path + "-" + str, ok
		}
	}
	return "", true
}
