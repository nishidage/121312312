package archive

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"arhat.dev/pkg/fshelper"
	"arhat.dev/pkg/sorthelper"
)

type entry struct {
	to   string
	from string
	info fs.FileInfo

	link string
}

// collectFiles to be archived
func collectFiles(
	ofs *fshelper.OSFS,
	files []*archiveFileSpec,
) ([]*entry, error) {
	var (
		swap []*entry

		inArchiveFiles = make(map[string]int)

		ret []*entry
	)

	for _, f := range files {
		srcPaths, err := ofs.Glob(f.From)
		if err != nil {
			srcPaths = []string{f.From}
		}

		// normalize matched paths
		slashMatches := make([]string, len(srcPaths))
		for i, v := range srcPaths {
			if len(v) == 0 {
				v = "."
			}

			info, err := ofs.Lstat(v)
			if err != nil {
				return nil, err
			}

			link := ""
			if info.Mode()&fs.ModeSymlink != 0 {
				link, err = ofs.Readlink(v)
				if err != nil {
					return nil, err
				}
			}

			swap = append(swap, &entry{
				info: info,
				from: v,
				link: link,
			})

			slashMatches[i] = filepath.ToSlash(v)
			if info.IsDir() {
				slashMatches[i] += "/"
			}
		}

		size := len(swap)

		dstIsDir := strings.HasSuffix(f.To, "/") || len(f.To) == 0
		switch len(slashMatches) {
		case 0:
			return nil, fmt.Errorf("no file matches pattern %q", f.From)
		case 1:
			toAdd := swap[size-1]
			// only one file match
			if dstIsDir {
				toAdd.to = path.Join(f.To, filepath.Base(toAdd.from))
			} else {
				toAdd.to = f.To
			}

			inArchiveFiles[toAdd.to] = len(ret)
			ret = append(ret, toAdd)
			continue
		default:
		}

		// multiple file matches, expect dir for these files
		if !dstIsDir {
			return nil, fmt.Errorf("too many files to %q", f.To)
		}

		prefix := lcpp(slashMatches)
		for i, slashPath := range slashMatches {
			if slashPath == prefix {
				continue
			}

			toAdd := swap[size-len(slashMatches)+i]
			toAdd.to = path.Join(f.To, strings.TrimPrefix(slashPath, prefix))

			inArchiveFiles[toAdd.to] = len(ret)
			ret = append(ret, toAdd)
		}
	}

	// add missing directories
	good := make(map[string]struct{})
	lastKnownGood := -1
	for lastKnownGood != len(good) {
		lastKnownGood = len(good)
		for k, idx := range inArchiveFiles {
			dir := path.Dir(k)
			if dir == "." || dir == "/" {
				good[k] = struct{}{}
				continue
			}

			_, ok1 := inArchiveFiles[dir]
			_, ok2 := inArchiveFiles[dir+"/"]
			if ok1 || ok2 {
				good[k] = struct{}{}
				continue
			}

			// no parent dir, add a fake one based on
			// actual parent of the file
			from := filepath.Dir(ret[idx].from)
			info, err := ofs.Lstat(from)
			if err != nil {
				return nil, err
			}

			ent := &entry{
				from: from,
				info: info,
				to:   dir,
				link: "",
			}

			inArchiveFiles[dir] = len(ret)
			ret = append(ret, ent)
		}
	}

	sort.Sort(sorthelper.NewCustomSortable(
		func(i, j int) {
			ret[i], ret[j] = ret[j], ret[i]
		}, func(i, j int) bool {
			return ret[i].to < ret[j].to
		}, func() int {
			return len(ret)
		},
	))

	return ret, nil
}

func copyFileContent(ofs *fshelper.OSFS, w io.Writer, file string) error {
	f, err := ofs.Open(file)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}

	return nil
}

// longest common path prefix
func lcpp(l []string) string {
	if len(l) == 0 {
		return ""
	}

	min, max := l[0], l[0]
	for _, s := range l[1:] {
		switch {
		case s < min:
			min = s
		case s > max:
			max = s
		}
	}

	lastSlashAt := -1
	for i := 0; i < len(min) && i < len(max); i++ {
		if min[i] != max[i] {
			break
		}

		if min[i] == '/' {
			lastSlashAt = i
		}
	}

	if lastSlashAt != -1 {
		return min[:lastSlashAt+1]
	}

	return ""
}
