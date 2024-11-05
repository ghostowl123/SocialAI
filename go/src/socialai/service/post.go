package service

import (
	"mime/multipart"
	"reflect"
	"socialai/backend"
	"socialai/constants"
	"socialai/model"

	"github.com/olivere/elastic/v7"
)

func SearchPostsByUser(user string) ([]model.Post, error) {
	// 1. create a query
	query := elastic.NewTermQuery("user", user)

	// 2. call backend
	searchResult, err := backend.ESBackend.ReadFromES(query, constants.POST_INDEX)
	if err != nil {
		return nil, err
	}

	return getPostFromSearchResult(searchResult), nil
}

func SearchPostsByKeywords(keywords string) ([]model.Post, error) {
	//option1:return nothing
	// if keywords == "" {
	//  return nil, nil
	// }

	// 1. create a query
	query := elastic.NewMatchQuery("message", keywords)
	query.Operator("AND")
	//option 2, return all
	if keywords == "" {
		query.ZeroTermsQuery("all")
	}

	// 2. call backend
	searchResult, err := backend.ESBackend.ReadFromES(query, constants.POST_INDEX)
	if err != nil {
		return nil, err
	}

	return getPostFromSearchResult(searchResult), nil
}

func getPostFromSearchResult(searchResult *elastic.SearchResult) []model.Post {
	var ptype model.Post
	var posts []model.Post

	for _, item := range searchResult.Each(reflect.TypeOf(ptype)) {
		p := item.(model.Post)
		posts = append(posts, p)
	}
	return posts
}

func SavePost(post *model.Post, file multipart.File) error {
	medialink, err := backend.GCSBackend.SaveToGCS(file, post.Id)
	if err != nil {
		return err
	}
	post.Url = medialink
	//SaveToEs 可能会出现问题 GCS 存进去了 ES没有 非原子性操作：得要么同时成功要么同时失败
	return backend.ESBackend.SaveToES(post, constants.POST_INDEX, post.Id)
	// 怎么处理？
	//if err!= nil{
	//     1.retury
	//     2.rollback
	//     3.backend.GCSBackend.DeleteFromGCS
	// }
	// 另一类方法：
	// 设置程序 定期清理我们的GCS里面有没有没用的 ES里面没有的

}
func DeletePost(id string, user string) error {
	query := elastic.NewBoolQuery()
	query.Must(elastic.NewTermQuery("id", id))
	query.Must(elastic.NewTermQuery("user", user))

	return backend.ESBackend.DeleteFromES(query, constants.POST_INDEX)
}
