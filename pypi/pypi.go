package pypi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
	endpoint := fmt.Sprintf("pypi/%s/%s/json", projectName, version)
	pkg, err = p.packageReq(endpoint)
	return pkg, nil
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
