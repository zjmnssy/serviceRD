package detector

import "google.golang.org/grpc/resolver"

func getDataFromMeta(addr resolver.Address, key string) (string, bool) {
	var ok bool
	var metadata *map[string]string
	var data string

	if addr.Metadata != nil {
		metadata, ok = addr.Metadata.(*map[string]string)
		if ok {
			data, ok = (*metadata)[key]
			if !ok {
				return data, false
			}
		} else {
			return data, false
		}
	}

	return data, true
}
