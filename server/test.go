package main

import (
	"fmt"

	art "github.com/plar/go-adaptive-radix-tree/v2"
)

func main() {
	tree := art.New()

	tree.Insert(art.Key("/foo/bar"), 1)
	tree.Insert(art.Key("/foo"), 1)
	tree.Insert(art.Key("/abc"), 1)
	tree.Insert(art.Key("/abc/def"), 1)
	tree.Insert(art.Key("/abc/def/ghi"), 1)
	tree.Insert(art.Key("/etc"), 1)

	tree.ForEach(func(node art.Node) (cont bool) {
		fmt.Println("node:", string(node.Key()))

		return true
	})

	fmt.Println("-----")

	tree.ForEachPrefix(art.Key("/"), func(node art.Node) (cont bool) {
		fmt.Println("node:", string(node.Key()))

		return true
	}, art.TraverseLeaf)

	tree.Iterator()
}
