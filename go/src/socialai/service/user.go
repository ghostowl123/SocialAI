package service

import (
	"fmt"
	"reflect"
	"socialai/backend"
	"socialai/constants"
	"socialai/model"

	"github.com/olivere/elastic/v7"
)

func CheckUser(username, password string) (bool, error) {
	// 1. read Search ES by username, then compare passwords
	// 2. Search ES by username + password , TotalHit() > 0
	//下面 相当于 1+2 先去ES里面验证了一下 有 之后读取结果再对比一遍
	query := elastic.NewBoolQuery()
	query.Must(elastic.NewTermQuery("username", username))
	query.Must(elastic.NewTermQuery("password", password))
	searchResult, err := backend.ESBackend.ReadFromES(query, constants.USER_INDEX)
	if err != nil {
		return false, err
	}

	var utype model.User
	for _, item := range searchResult.Each(reflect.TypeOf(utype)) {
		u := item.(model.User)
		if u.Password == password {
			fmt.Printf("Login as %s\n", username)
			return true, nil
		}
	}
	return false, nil
}

// 如果 mySQL 不需要 因为unique key 如果有重复那么就报错
// 如果 noSQL unique key 如果有重复会overwrite数据 这样我们就需要一个additional bool 去判断数据是否存在
func AddUser(user *model.User) (bool, error) {
	//1. check whether username existed
	query := elastic.NewTermQuery("username", user.Username)
	searchResult, err := backend.ESBackend.ReadFromES(query, constants.USER_INDEX)
	if err != nil {
		return false, err
	}

	if searchResult.TotalHits() > 0 {
		return false, nil
	}
    // if not, add user
	err = backend.ESBackend.SaveToES(user, constants.USER_INDEX, user.Username)
	if err != nil {
		return false, err
	}
	fmt.Printf("User is added: %s\n", user.Username)
	return true, nil
}
