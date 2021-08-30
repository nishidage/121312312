package rs

import (
	"encoding/json"
	"fmt"
	"reflect"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"gopkg.in/yaml.v3"
)

type MergeSource struct {
	BaseField `yaml:"-" json:"-"`

	Data *AnyObject `yaml:"data,omitempty"`
}

type resolvedMergeSource struct {
	Data interface{} `yaml:"data,omitempty"`
}

type renderingPatchSpec struct {
	BaseField

	// Value for the renderer
	//
	// 	say we have a yaml list ([bar]) stored at https://example.com/bar.yaml
	//
	// 		foo@http!:
	// 		  value: https://example.com/bar.yaml
	// 		  merge: { data: [foo] }
	//
	// then the resolve value of foo will be [bar, foo]
	Value *AnyObject `yaml:"value"`

	// Merge additional data into Value
	Merge []MergeSource `yaml:"merge,omitempty"`

	// Patches Value using standard rfc6902 json-patch
	Patches []JSONPatchSpec `yaml:"patches"`

	// Unique to make sure elements in the sequence is unique
	//
	// only effective when Value is yaml sequence
	Unique bool `yaml:"unique"`

	// MapListItemUnique to ensure items are unique in all merged lists respectively
	// lists with no merge data input are untouched
	MapListItemUnique bool `yaml:"map_list_item_unique"`

	// MapListAppend to append lists instead of replacing existing list
	MapListAppend bool `yaml:"map_list_append"`
}

func (s *renderingPatchSpec) merge(resolvedValueData, resolvedMergeData []byte) (interface{}, error) {
	var valueData interface{}
	if len(resolvedValueData) != 0 {
		err := yaml.Unmarshal(resolvedValueData, &valueData)
		if err != nil {
			return nil, err
		}
	}

	var mergeSrc []resolvedMergeSource
	if len(resolvedMergeData) != 0 {
		err := yaml.Unmarshal(resolvedMergeData, &mergeSrc)
		if err != nil {
			return nil, err
		}
	}

doMerge:
	switch dt := valueData.(type) {
	case []interface{}:
		for _, merge := range mergeSrc {
			switch mt := merge.Data.(type) {
			case []interface{}:
				dt = append(dt, mt...)

				if s.Unique {
					dt = UniqueList(dt)
				}
			case nil:
				// no value to merge, skip
			default:
				// invalid type, not able to merge
				return nil, fmt.Errorf("unexpected non list value of merge, got %T", mt)
			}
		}

		return dt, nil
	case map[string]interface{}:
		var err error
		for _, merge := range mergeSrc {
			switch mt := merge.Data.(type) {
			case map[string]interface{}:
				dt, err = MergeMap(dt, mt, s.MapListAppend, s.MapListItemUnique)
				if err != nil {
					return nil, fmt.Errorf("failed to merge map value: %w", err)
				}
			case nil:
				// no value to merge, skip
			default:
				// invalid type, not able to merge
				return nil, fmt.Errorf("unexpected non map value of merge, got %T", mt)
			}
		}

		return dt, nil
	case nil:
		// TODO: do we really want to allow empty value?
		// 		 could it be some kind of error that should be checked during merging?
		switch len(mergeSrc) {
		case 0:
			return nil, nil
		case 1:
			return mergeSrc[0].Data, nil
		default:
			valueData = mergeSrc[0].Data
			mergeSrc = mergeSrc[1:]
			goto doMerge
		}
	default:
		// scalar types, not supported
		return nil, fmt.Errorf("mergering scalar type value is not supported")
	}
}

// Apply Merge and Patch to Value, Unique is ensured if set to true
func (s *renderingPatchSpec) ApplyTo(resolvedValueData, resolvedMergeData []byte) ([]byte, error) {
	data, err := s.merge(resolvedValueData, resolvedMergeData)
	if err != nil {
		return nil, err
	}

	if len(s.Patches) == 0 {
		return yaml.Marshal(data)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	patchData, err := json.Marshal(s.Patches)
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.DecodePatch(patchData)
	if err != nil {
		return nil, err
	}

	return patch.ApplyIndentWithOptions(jsonData, "", &jsonpatch.ApplyOptions{
		SupportNegativeIndices:   true,
		EnsurePathExistsOnAdd:    true,
		AccumulatedCopySizeLimit: 0,
		AllowMissingPathOnRemove: false,
	})
}

func MergeMap(
	original, additional map[string]interface{},

	// options
	appendList bool,
	uniqueInListItems bool,
) (map[string]interface{}, error) {
	out := make(map[string]interface{}, len(original))
	for k, v := range original {
		out[k] = v
	}

	var err error
	for k, v := range additional {
		switch newVal := v.(type) {
		case map[string]interface{}:
			if originalVal, ok := out[k]; ok {
				if orignalMap, ok := originalVal.(map[string]interface{}); ok {
					out[k], err = MergeMap(orignalMap, newVal, appendList, uniqueInListItems)
					if err != nil {
						return nil, err
					}

					continue
				} else {
					return nil, fmt.Errorf("unexpected non map data %q: %v", k, orignalMap)
				}
			} else {
				out[k] = newVal
			}
		case []interface{}:
			if originalVal, ok := out[k]; ok {
				if originalList, ok := originalVal.([]interface{}); ok {
					if appendList {
						originalList = append(originalList, newVal...)
					} else {
						originalList = newVal
					}

					if uniqueInListItems {
						originalList = UniqueList(originalList)
					}

					out[k] = originalList

					continue
				} else {
					return nil, fmt.Errorf("unexpected non list data %q: %v", k, originalList)
				}
			} else {
				out[k] = newVal
			}
		default:
			out[k] = newVal
		}
	}

	return out, nil
}

func UniqueList(dt []interface{}) []interface{} {
	var ret []interface{}
	dupAt := make(map[int]struct{})
	for i := range dt {
		_, isDup := dupAt[i]
		if isDup {
			continue
		}

		for j := i; j < len(dt); j++ {
			if reflect.DeepEqual(dt[i], dt[j]) {
				dupAt[j] = struct{}{}
			}
		}

		ret = append(ret, dt[i])
	}

	return ret
}

// JSONPatchSpec per rfc6902
type JSONPatchSpec struct {
	BaseField `yaml:"-" json:"-"`

	Operation string `yaml:"op" json:"op"`

	Path string `yaml:"path" json:"path"`

	Value interface{} `yaml:"value,omitempty" json:"value,omitempty"`
}
