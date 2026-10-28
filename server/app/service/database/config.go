package database

import (
	"encoding/json"
	"errors"
	"reflect"
	"server/app/enum/database_kind"
	"server/app/service/database/elasticsearch"
	"server/app/service/database/mongo"
	"server/app/service/database/mysql"
	"server/app/service/database/postgres"
	"server/app/service/database/redis"
	"server/app/service/database/sqlite"
	"server/app/service/database/sqlserver"
	"strconv"
	"strings"
)

// Config 解析连接配置，按 encrypt:"true" 标签处理凭据；action 为空时仅解析，encrypt 加密返回，decrypt 解密提交。
func (s *service) Config(databaseType int, content string, id int64, stored string, action string) (interface{}, error) {
	if action != "" && action != "encrypt" && action != "decrypt" {
		return nil, errors.New("配置处理方式不合法")
	}
	var config interface{}
	switch databaseType {
	case database_kind.MySQL:
		config = mysql.Config{}
	case database_kind.Postgres:
		config = postgres.Config{}
	case database_kind.SqlServer:
		config = sqlserver.Config{}
	case database_kind.Sqlite:
		config = sqlite.Config{}
	case database_kind.Redis:
		config = redis.Config{}
	case database_kind.MongoDB:
		config = mongo.Config{}
	case database_kind.Elasticsearch:
		config = elasticsearch.Config{}
	default:
		return nil, errors.New("数据库类型不合法")
	}
	content = strings.TrimSpace(content)
	// 首次配置采用各数据库推荐的连接方式，已保存的 TLS 选项保持原意。
	if content == "" {
		switch databaseType {
		case database_kind.Elasticsearch:
			content = `{"tls":"true"}`
		case database_kind.MySQL:
			content = `{"tls":"preferred"}`
		}
	}
	values := make(map[string]json.RawMessage)
	if content != "" && (len(content) > 1<<20 || json.Unmarshal([]byte(content), &values) != nil || values == nil) {
		return nil, errors.New("连接配置必须是 JSON 对象，且不超过 1 MB")
	}
	value := reflect.New(reflect.TypeOf(config)).Elem()
	definition := value.Type()
	for i := 0; i < definition.NumField(); i++ {
		property := definition.Field(i)
		name := strings.Split(property.Tag.Get("json"), ",")[0]
		label := property.Tag.Get("label")
		raw := values[name]
		delete(values, name)
		target := value.Field(i)
		text := ""
		if len(raw) != 0 && string(raw) != "null" {
			if json.Unmarshal(raw, &text) != nil {
				if raw[0] == '{' || raw[0] == '[' || (target.Kind() == reflect.String &&
					(property.Tag.Get("options") == "" || (string(raw) != "true" && string(raw) != "false"))) {
					return nil, errors.New(label + "格式不正确")
				}
				text = string(raw)
			}
		}
		if text == "" {
			text = property.Tag.Get("default")
		}
		switch target.Kind() {
		case reflect.String:
			if action != "" && property.Tag.Get("encrypt") == "true" && text != "" {
				// 凭据密文绑定实例、字段和已保存配置，避免被其他实例复用。
				key := "database-config:" + strconv.FormatInt(id, 10) + ":" + strconv.Itoa(databaseType) + ":" + name + ":" + stored
				if action == "encrypt" {
					if id <= 0 || stored == "" {
						return nil, errors.New("实例连接配置不存在")
					}
					data := s.encrypt(key, text)
					if data == "" {
						return nil, errors.New("连接凭据加密失败")
					}
					text = "enc:v1:" + data
				} else if data, ok := strings.CutPrefix(text, "enc:v1:"); ok {
					text = s.decrypt(key, data)
					if id <= 0 || stored == "" || text == "" {
						return nil, errors.New("连接凭据已失效，请重新加载实例信息")
					}
				}
			}
			target.SetString(text)
		case reflect.Int:
			number, err := strconv.ParseInt(text, 10, target.Type().Bits())
			if err != nil {
				return nil, errors.New(label + "必须是整数")
			}
			if minimum, err := strconv.ParseInt(property.Tag.Get("min"), 10, 64); err == nil && number < minimum {
				return nil, errors.New(label + "不能小于 " + strconv.FormatInt(minimum, 10))
			}
			if maximum, err := strconv.ParseInt(property.Tag.Get("max"), 10, 64); err == nil && number > maximum {
				return nil, errors.New(label + "不能大于 " + strconv.FormatInt(maximum, 10))
			}
			target.SetInt(number)
		case reflect.Bool:
			boolean, err := strconv.ParseBool(text)
			if err != nil {
				return nil, errors.New(label + "必须是布尔值")
			}
			target.SetBool(boolean)
		}
		if property.Tag.Get("required") == "true" && content != "" && strings.TrimSpace(text) == "" {
			return nil, errors.New("请填写" + label)
		}
		if options := property.Tag.Get("options"); options != "" {
			matched := false
			for _, option := range strings.Split(options, ",") {
				key, _, _ := strings.Cut(option, ":")
				if key == strings.ToLower(text) {
					matched = true
					target.SetString(key)
				}
			}
			if !matched {
				return nil, errors.New(label + "选项不合法")
			}
		}
	}
	for _, raw := range values {
		if string(raw) != `""` && string(raw) != "null" {
			return nil, errors.New("连接配置包含不支持的字段，请分别填写各项连接参数")
		}
	}
	return value.Interface(), nil
}
