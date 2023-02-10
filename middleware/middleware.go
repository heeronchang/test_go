// net/http 中间件
package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

type MiddleWare func(http.Handler) http.Handler

func ApplyMiddlewares(handler http.Handler, middles ...MiddleWare) http.Handler {
	for i := len(middles); i >= 0; i-- {
		handler = middles[i](handler)
	}

	return handler
}

// Metric collect elapse time
func Metirc(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		defer func() {
			log.Printf("path:%s elapsed:%fs\n", r.URL.Path, time.Since(t).Seconds())
		}()
		time.Sleep(time.Second * 2)
		handler.ServeHTTP(w, r)
	})
}

// PanicRecover 处理错误恢复
func PanicRecover(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(string(debug.Stack()))
			}
		}()

		handler.ServeHTTP(w, r)
	})
}

// WithLogger print log
func WithLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("path:%s process start..\n", r.URL.Path)
		defer func() {
			log.Printf("path:%s process end.\n", r.URL.Path)
		}()
		handler.ServeHTTP(w, r)
	})
}
