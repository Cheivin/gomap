package gomap

type (
	Map interface {
		Store(key string, value interface{})                                                                           // 存储key-val
		Load(key string) (value interface{}, ok bool)                                                                  // 查找key-val
		LoadOrStore(key string, value interface{}) (actual interface{}, loaded bool)                                   // 查找key-val，存在则返回原有值，不存在则放入新值返回
		StoreOrCompare(key string, value interface{}, compare func(stored interface{}, input interface{}) interface{}) // 比较并存储compare返回值
		Delete(key string) interface{}                                                                                 // 删除指定key，成功返回被删除val
		Clear() []Entry                                                                                                // 清空
		Range(f func(key, value interface{}) bool)                                                                     // 遍历
		Destroy()                                                                                                      // 销毁
		Size() int                                                                                                     // 大小
	}
	Entry struct {
		Key   string
		Value interface{}
	}
)

const ErrMapDestroyed = "ErrMapDestroyed"
