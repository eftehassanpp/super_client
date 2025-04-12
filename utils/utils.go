package utils

import (
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

func MakeFormData(data map[string]any) string {
	formData := url.Values{}
	for key, value := range data {
		switch v := value.(type) {
		case string:
			formData.Add(key, v)
		case int:
			formData.Add(key, strconv.Itoa(v))
		case bool:
			formData.Add(key, strconv.FormatBool(v))
		case map[string]string:
			// Handle map for "selected_payment"
			for subKey, subValue := range v {
				formData.Add(key+"["+subKey+"]", subValue)
			}
		}
	}

	// Print the URL-encoded form data
	encodedFormData := formData.Encode()
	// fmt.Println(encodedFormData)
	return encodedFormData
}

func GenerateRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}
