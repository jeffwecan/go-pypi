package pypi

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cavaliercoder/grab"
)

type PackageIndex struct {
	url    string
	Client http.Client
}

type Package struct {
	Info       Info                 `json:"info"`
	LastSerial int                  `json:"last_serial"`
	Releases   map[string][]Release `json:"releases"`
	Urls       []Release            `json:"urls"`
}

type Info struct {
	Author                 string        `json:"author"`
	AuthorEmail            string        `json:"author_email"`
	BugtrackUrl            string        `json:"bugtrack_url"`
	Classifiers            []string      `json:"classifiers"`
	Description            string        `json:"description"`
	DescriptionContentType string        `json:"description_content_type"`
	DocsUrl                string        `json:"docs_url"`
	DownloadUrl            string        `json:"download_url"`
	Downloads              InfoDownloads `json:"downloads"`
	HomePage               string        `json:"home_page"`
	Keywords               string        `json:"keywords"`
	License                string        `json:"license"`
	Maintainer             string        `json:"maintainer"`
	MaintainerEmail        string        `json:"maintainer_email"`
	Name                   string        `json:"name"`
	PackageUrl             string        `json:"package_url"`
	Platform               string        `json:"platform"`
	ProjectUrl             string        `json:"project_url"`
	ReleaseUrl             string        `json:"release_url"`
	RequiresDist           string        `json:"requires_dist"`
	RequiresPython         string        `json:"requires_python"`
	Summary                string        `json:"summary"`
	Version                string        `json:"version"`
	Yanked                 bool          `json:"yanked"`
	YankedReason           string        `json:"yanked_reason"`
}

type InfoDownloads struct {
	LastDay   int `json:"last_day"`
	LastMonth int `json:"last_month"`
	LastWeek  int `json:"last_week"`
}

type Release struct {
	CommentText       string         `json:"comment_text"`
	Digest            ReleaseDigests `json:"digests"`
	Downloads         int            `json:"downloads"`
	Filename          string         `json:"filename"`
	HasSig            bool           `json:"has_sig"`
	Md5Digest         string         `json:"md5_digest"`
	PackageType       string         `json:"packagetype"`
	PythonVersion     string         `json:"python_version"`
	Size              int            `json:"size"`
	UploadTimeIso8601 string         `json:"upload_time_iso_8601"`
	Url               string         `json:"url"`
	Yanked            bool           `json:"yanked"`
	YankedReason      string         `json:"yanked_reason"`
}

type ReleaseDigests struct {
	Md5    string `json:"md5"`
	Sha256 string `json:"sha256"`
}

func NewPackageIndex(url string) *PackageIndex {
	p := &PackageIndex{
		url: url,
	}
	p.Client = http.Client{}
	return p
}

func (p *Package) GetWheelByVersion(version string) (wheel Release) {
	for _, release := range p.Releases[version] {
		log.Printf("v%s: %s", version, release.PackageType)
		if release.PackageType == "bdist_wheel" {
			wheel = release
		}
	}
	return wheel
}

func (p *Package) GetSdistByVersion(version string) (sdist Release) {
	for _, release := range p.Releases[version] {
		if release.PackageType == "sdist" {
			sdist = release
		}
	}
	return sdist
}

func (p *PackageIndex) packageReq(endpoint string) (pkg Package, err error) {
	url := fmt.Sprintf("%s/%s", p.url, endpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return pkg, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.Client.Do(req)
	if err != nil {
		return pkg, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	// log.Print(string(data))
	log.Print("gots some datas probably")
	if err != nil {
		return pkg, err
	}
	err = json.Unmarshal(data, &pkg)
	if err != nil {
		return pkg, err
	}
	return pkg, err

}

func (p *PackageIndex) GetLatest(projectName string) (pkg Package, err error) {
	endpoint := fmt.Sprintf("pypi/%s/json", projectName)
	pkg, err = p.packageReq(endpoint)
	return pkg, nil
}

func (p *PackageIndex) GetRelease(projectName string, version string) (pkg Package, err error) {
	return p.GetLatest(projectName)
}

func downloadReleaseFile(dst, url string) error {
	client := grab.NewClient()
	req, _ := grab.NewRequest(dst, url)

	log.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	log.Printf("  %v\n", resp.HTTPResponse.Status)
	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			log.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		return errors.New(fmt.Sprintf("downloading release file failed: %v\n", err))
	}

	// fmt.Printf("Download saved to ./%v \n", resp.Filename)
	// check for errors,... here,... somehow?
	// if err != nil {
	// 	log.Printf("Download failed: %v\n", err)
	// 	return err
	// }

	log.Printf("Download saved to ./%v \n", resp.Filename)
	return nil
}

func (p *PackageIndex) DownloadLatest(dst, projectName string) (filename string, err error) {
	pkg, err := p.GetLatest(projectName)
	if err != nil {
		return filename, errors.New(fmt.Sprintf("error downloading latest: %s", err))
	}
	filename, err = p.DownloadRelease(dst, projectName, pkg.Info.Version)
	return filename, err
}

func (p *PackageIndex) DownloadRelease(dst, projectName string, version string) (filename string, err error) {
	pkg, err := p.GetRelease(projectName, version)
	if err != nil {
		return filename, errors.New(fmt.Sprintf("error getting release: %s", err))
	}
	release := pkg.GetWheelByVersion(version)
	if release.Url == "" {
		release = pkg.GetSdistByVersion(version)
	}
	if release.Url == "" {
		return release.Filename, errors.New(fmt.Sprintf("no release found to download for %s v%s", projectName, version))
	}
	err = downloadReleaseFile(dst, release.Url)
	return release.Filename, err
}

func (p *PackageIndex) DownloadFromRequirementsFile(dst, filename string) (reqs []Requirement, err error) {
	reqs, err = ParseRequirements(filename)
	for _, req := range reqs {
		if req.Specification.Comparison == "==" {
			log.Printf("about to download: %s", req.Name)
			filename, err = p.DownloadRelease(dst, req.Name, req.Specification.Version)
			if err != nil {
				return reqs, err
			}
			extension := filepath.Ext(filename)
			var extractedFiles []string

			if extension == ".gz" {
				log.Printf("about to untar: %s", filename)
				r, err := os.Open(filepath.Join(dst, filename))
				if err != nil {
					return reqs, fmt.Errorf("error opening file at: %s", filepath.Join(dst, filename))
				}
				extractedFiles, err = Untar(dst, r)
				if err != nil {
					return reqs, fmt.Errorf("error extracting requirement file %s: %s", filename, err)
				}
				// fmt.Printf("untar of %s files: %+v", filename, extractedFiles)

			} else {
				log.Printf("about to unzip: %s", filename)
				extractedFiles, err = Unzip(filepath.Join(dst, filename), dst)
				if err != nil {
					return reqs, fmt.Errorf("error unzipping requirement file %s: %s", filename, err)
				}
				// log.Printf("unzipped:\n" + strings.Join(extractedFiles, "\n"))

			}

			for _, extractedFile := range extractedFiles {
				extractDirName := fmt.Sprintf("/%s", strings.ReplaceAll(filename, ".tar.gz", ""))
				oldLocation := extractedFile
				newLocation := strings.ReplaceAll(oldLocation, extractDirName, "")
				log.Printf("mv %s ==> %s", oldLocation, newLocation)
				if oldLocation == newLocation {
					log.Printf("not moving this deal: old path (%s) == new path (%s)", oldLocation, newLocation)
					continue
				}
				if _, err := os.Stat(newLocation); err == nil {
					fmt.Printf("Not moving %s; already present at: %s", oldLocation, newLocation)
				} else if errors.Is(err, os.ErrNotExist) {
					err = os.Rename(oldLocation, newLocation)
					// if  .
					if err != nil && !os.IsExist(err) {
						return reqs, fmt.Errorf("error moving module directory after extracting %s: %s", filename, err)
					}
				} else {
					log.Printf("not moving this deal: %s", newLocation)
				}

			}

		} else {
			log.Panicf("unsure how to deal with this dude: %+v", req)
		}
	}
	return reqs, err
}

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func Untar(dst string, r io.Reader) ([]string, error) {

	var filenames []string
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return filenames, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return filenames, nil

		// return any other error
		case err != nil:
			return filenames, err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)
		filenames = append(filenames, target)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return filenames, err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return filenames, err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return filenames, err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
	// return filenames, nil
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		// if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
		//     return filenames, fmt.Errorf("%s: illegal file path", fpath)
		// }

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
