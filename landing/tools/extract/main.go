// Découpe les assets landing (chroma + blobs du splash) et écrit coin.glb.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	root := findLandingRoot()
	raw := filepath.Join(root, "raw")
	out := filepath.Join(root, "assets")
	must(os.MkdirAll(out, 0755))
	must(os.MkdirAll(filepath.Join(out, "from-splash"), 0755))

	type job struct {
		src, dst string
		fn       func(image.Image) *image.NRGBA
	}
	jobs := []job{
		{"asset-card-pistolets.png", "card-pistolets.png", keyMagenta},
		{"asset-card-style.png", "card-style.png", keyMagenta},
		{"asset-card-ready.png", "card-ready.png", keyMagenta},
		{"asset-btn-download.png", "btn-download.png", keyMagenta},
		{"asset-btn-login.png", "btn-login.png", keyMagenta},
		{"asset-spell-q.png", "spell-q.png", keyMagenta},
		{"asset-spell-w.png", "spell-w.png", keyMagenta},
		{"asset-spell-e.png", "spell-e.png", keyMagenta},
		{"asset-spell-r.png", "spell-r.png", keyMagenta},
		{"asset-icon-arme.png", "icon-arme.png", keyMagenta},
		{"asset-icon-style.png", "icon-style.png", keyMagenta},
		{"asset-icon-allin.png", "icon-allin.png", keyMagenta},
		{"asset-coin-front.png", "coin-front.png", coinCircle},
		{"asset-coin-back.png", "coin-back.png", coinCircle},
		{"asset-smoke.png", "smoke.png", keyBlackSoft},
		{"asset-glass-crack.png", "glass-crack.png", keyBlackSoft},
		{"asset-revolver-ref.png", "revolver-ref.png", keyBlackSoft},
		{"splash.png", "splash.png", copyNRGBA},
	}
	for _, j := range jobs {
		im, err := loadPNG(filepath.Join(raw, j.src))
		if err != nil {
			fmt.Println("skip", j.src, err)
			continue
		}
		cut := tight(keepLargest(j.fn(im), 0.008))
		must(savePNG(filepath.Join(out, j.dst), cut))
		fmt.Printf("ok %s %dx%d\n", j.dst, cut.Bounds().Dx(), cut.Bounds().Dy())
	}

	splash, err := loadPNG(filepath.Join(raw, "splash.png"))
	if err != nil {
		panic(err)
	}
	extractSplash(splash, filepath.Join(out, "from-splash"))

	front, err := loadPNG(filepath.Join(out, "coin-front.png"))
	must(err)
	back, err := loadPNG(filepath.Join(out, "coin-back.png"))
	must(err)
	glb := buildCoinGLB(front, back)
	must(os.WriteFile(filepath.Join(out, "coin.glb"), glb, 0644))
	fmt.Printf("ok coin.glb %d bytes\n", len(glb))
}

func findLandingRoot() string {
	wd, _ := os.Getwd()
	cands := []string{wd, filepath.Join(wd, "landing"), filepath.Dir(wd), filepath.Join(filepath.Dir(wd), "landing")}
	for _, c := range cands {
		if st, err := os.Stat(filepath.Join(c, "raw")); err == nil && st.IsDir() {
			return c
		}
	}
	return filepath.Join(wd, "landing")
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	im, err := png.Decode(f)
	return im, err
}

func savePNG(path string, im image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	return enc.Encode(f, im)
}

func toNRGBA(im image.Image) *image.NRGBA {
	if n, ok := im.(*image.NRGBA); ok {
		return n
	}
	b := im.Bounds()
	out := image.NewNRGBA(b)
	draw.Draw(out, b, im, b.Min, draw.Src)
	return out
}

func copyNRGBA(im image.Image) *image.NRGBA { return toNRGBA(im) }

func keyMagenta(im image.Image) *image.NRGBA {
	src := toNRGBA(im)
	b := src.Bounds()
	out := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := src.NRGBAAt(x, y)
			a := magentaAlpha(i)
			i.A = a
			if a == 0 {
				i.R, i.G, i.B = 0, 0, 0
			} else if a < 255 {
				// un-spill magenta
				i.R = uint8(int(i.R) * int(a) / 255)
				i.B = uint8(int(i.B) * int(a) / 255)
			}
			out.SetNRGBA(x, y, i)
		}
	}
	return out
}

func magentaAlpha(c color.NRGBA) uint8 {
	mg := int(c.G)
	rb := (int(c.R) + int(c.B)) / 2
	diff := absInt(int(c.R) - int(c.B))
	md := (255-int(c.R))*(255-int(c.R)) + mg*mg + (255-int(c.B))*(255-int(c.B))
	if md < 55*55 && mg < 90 {
		return 0
	}
	if rb > 150 && mg < 110 && diff < 55 && mg < rb*2/5 {
		t := float64(110-mg) / 110
		if t > 1 {
			t = 1
		}
		a := 255 * (1 - t*t)
		if a < 0 {
			a = 0
		}
		return uint8(a)
	}
	return 255
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func keepLargest(im *image.NRGBA, minFrac float64) *image.NRGBA {
	b := im.Bounds()
	w, h := b.Dx(), b.Dy()
	seen := make([]uint8, w*h)
	type pt struct{ x, y int }
	best := []pt{}
	q := make([]pt, 0, 1024)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			if seen[i] != 0 {
				continue
			}
			if im.NRGBAAt(b.Min.X+x, b.Min.Y+y).A < 12 {
				seen[i] = 1
				continue
			}
			q = q[:0]
			q = append(q, pt{x, y})
			seen[i] = 2
			comp := []pt{{x, y}}
			for len(q) > 0 {
				p := q[len(q)-1]
				q = q[:len(q)-1]
				for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := p.x+d[0], p.y+d[1]
					if nx < 0 || ny < 0 || nx >= w || ny >= h {
						continue
					}
					j := ny*w + nx
					if seen[j] != 0 {
						continue
					}
					if im.NRGBAAt(b.Min.X+nx, b.Min.Y+ny).A < 12 {
						seen[j] = 1
						continue
					}
					seen[j] = 2
					q = append(q, pt{nx, ny})
					comp = append(comp, pt{nx, ny})
				}
			}
			if len(comp) > len(best) {
				best = comp
			}
		}
	}
	if len(best) < int(float64(w*h)*minFrac) {
		return im
	}
	keep := make([]byte, w*h)
	for _, p := range best {
		keep[p.y*w+p.x] = 1
	}
	out := image.NewNRGBA(b)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := im.NRGBAAt(b.Min.X+x, b.Min.Y+y)
			if keep[y*w+x] == 0 {
				c.A = 0
			}
			out.SetNRGBA(b.Min.X+x, b.Min.Y+y, c)
		}
	}
	return out
}

func keyBlackSoft(im image.Image) *image.NRGBA {
	src := toNRGBA(im)
	b := src.Bounds()
	out := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := src.NRGBAAt(x, y)
			luma := (int(i.R)*3 + int(i.G)*6 + int(i.B)*1) / 10
			if luma < 12 {
				i.A = 0
			} else if luma < 40 {
				i.A = uint8((luma - 12) * 255 / 28)
			} else {
				i.A = 255
			}
			out.SetNRGBA(x, y, i)
		}
	}
	return out
}

func coinCircle(im image.Image) *image.NRGBA {
	src := keyBlackSoft(im)
	b := src.Bounds()
	// centroid of opaque pixels
	var sx, sy, n float64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if src.NRGBAAt(x, y).A > 80 {
				sx += float64(x)
				sy += float64(y)
				n++
			}
		}
	}
	if n < 100 {
		return src
	}
	cx, cy := sx/n, sy/n
	// radius = 92nd percentile of opaque distances
	dists := make([]float64, 0, int(n))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if src.NRGBAAt(x, y).A > 80 {
				dx := float64(x) - cx
				dy := float64(y) - cy
				dists = append(dists, math.Hypot(dx, dy))
			}
		}
	}
	sort.Float64s(dists)
	r := dists[int(float64(len(dists))*0.92)]
	out := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := src.NRGBAAt(x, y)
			d := math.Hypot(float64(x)-cx, float64(y)-cy)
			var a float64
			if d <= r-1.5 {
				a = 1
			} else if d >= r+1.5 {
				a = 0
			} else {
				a = 1 - (d-(r-1.5))/3
			}
			i.A = uint8(float64(i.A) * a)
			out.SetNRGBA(x, y, i)
		}
	}
	return out
}

func tight(im *image.NRGBA) *image.NRGBA {
	b := im.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X, b.Min.Y
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if im.NRGBAAt(x, y).A > 8 {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x >= maxX {
					maxX = x + 1
				}
				if y >= maxY {
					maxY = y + 1
				}
			}
		}
	}
	if maxX <= minX || maxY <= minY {
		return im
	}
	pad := 4
	minX = max(b.Min.X, minX-pad)
	minY = max(b.Min.Y, minY-pad)
	maxX = min(b.Max.X, maxX+pad)
	maxY = min(b.Max.Y, maxY+pad)
	r := image.Rect(0, 0, maxX-minX, maxY-minY)
	out := image.NewNRGBA(r)
	draw.Draw(out, r, im, image.Pt(minX, minY), draw.Src)
	return out
}

func extractSplash(im image.Image, dir string) {
	src := toNRGBA(im)
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()

	gold := blobs(src, func(c color.NRGBA) bool {
		return c.R > 165 && c.G > 105 && c.B < 95 && int(c.R) > int(c.G)+8 && int(c.G) > int(c.B)+25
	})
	teal := blobs(src, func(c color.NRGBA) bool {
		return c.G > 130 && c.B > 110 && c.R < 140 && int(c.G)+int(c.B) > int(c.R)*3+40
	})
	red := blobs(src, func(c color.NRGBA) bool {
		return c.R > 160 && c.G < 90 && c.B < 90 && int(c.R) > int(c.G)+70
	})

	coinN, spellN := 0, 0
	for _, bl := range gold {
		area := bl.dx() * bl.dy()
		if area < 180 || area > 14000 {
			continue
		}
		ar := float64(bl.dx()) / float64(bl.dy())
		if ar < 0.65 || ar > 1.45 {
			continue
		}
		if bl.minY < h/18 {
			continue
		}
		cut := cropFeather(src, bl, 6)
		coinN++
		must(savePNG(filepath.Join(dir, fmt.Sprintf("coin-%02d.png", coinN)), tight(cut)))
	}
	keys := append([]blob{}, teal...)
	keys = append(keys, red...)
	sort.Slice(keys, func(i, j int) bool { return keys[i].dx()*keys[i].dy() > keys[j].dx()*keys[j].dy() })
	for _, bl := range keys {
		area := bl.dx() * bl.dy()
		if area < 250 || area > 9000 {
			continue
		}
		ar := float64(bl.dx()) / float64(bl.dy())
		if ar < 0.7 || ar > 1.4 {
			continue
		}
		if bl.minY < h/14 || bl.minX < w/5 {
			continue
		}
		cut := cropFeather(src, bl, 8)
		spellN++
		must(savePNG(filepath.Join(dir, fmt.Sprintf("spell-%02d.png", spellN)), tight(cut)))
		if spellN >= 8 {
			break
		}
	}

	// cartes : plus gros rectangles teal / or / rouge hors header
	saveBiggest := func(list []blob, name string, minA, maxA int) {
		var best blob
		bestA := 0
		for _, bl := range list {
			a := bl.dx() * bl.dy()
			if a < minA || a > maxA {
				continue
			}
			if bl.minY < h/10 {
				continue
			}
			if a > bestA {
				bestA = a
				best = bl
			}
		}
		if bestA == 0 {
			return
		}
		cut := cropFeather(src, best, 10)
		must(savePNG(filepath.Join(dir, name), tight(cut)))
		fmt.Printf("splash %s %dx%d @%d,%d\n", name, best.dx(), best.dy(), best.minX, best.minY)
	}
	saveBiggest(teal, "card-pistolets.png", 8000, 90000)
	saveBiggest(gold, "card-style.png", 8000, 120000)
	saveBiggest(red, "card-ready.png", 6000, 90000)

	// CTA teal bas-gauche
	var cta blob
	ctaA := 0
	for _, bl := range teal {
		a := bl.dx() * bl.dy()
		if bl.minX > w/2 || bl.minY < h/3 || bl.minY > 4*h/5 {
			continue
		}
		ar := float64(bl.dx()) / float64(bl.dy())
		if ar < 1.8 || ar > 8 {
			continue
		}
		if a > ctaA && a > 1500 {
			ctaA = a
			cta = bl
		}
	}
	if ctaA > 0 {
		cut := cropFeather(src, cta, 8)
		must(savePNG(filepath.Join(dir, "btn-download.png"), tight(cut)))
		fmt.Printf("splash btn-download %dx%d\n", cta.dx(), cta.dy())
	}

	fmt.Printf("splash coins=%d spells=%d\n", coinN, spellN)
}

type blob struct{ minX, minY, maxX, maxY int }

func (b blob) dx() int { return b.maxX - b.minX }
func (b blob) dy() int { return b.maxY - b.minY }

func blobs(src *image.NRGBA, keep func(color.NRGBA) bool) []blob {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	seen := make([]byte, w*h)
	var out []blob
	type pt struct{ x, y int }
	q := make([]pt, 0, 4096)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			if seen[i] != 0 {
				continue
			}
			c := src.NRGBAAt(b.Min.X+x, b.Min.Y+y)
			if !keep(c) {
				seen[i] = 1
				continue
			}
			q = q[:0]
			q = append(q, pt{x, y})
			seen[i] = 2
			minX, minY, maxX, maxY := x, y, x, y
			n := 0
			for len(q) > 0 {
				p := q[len(q)-1]
				q = q[:len(q)-1]
				n++
				if p.x < minX {
					minX = p.x
				}
				if p.y < minY {
					minY = p.y
				}
				if p.x > maxX {
					maxX = p.x
				}
				if p.y > maxY {
					maxY = p.y
				}
				for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := p.x+d[0], p.y+d[1]
					if nx < 0 || ny < 0 || nx >= w || ny >= h {
						continue
					}
					j := ny*w + nx
					if seen[j] != 0 {
						continue
					}
					cc := src.NRGBAAt(b.Min.X+nx, b.Min.Y+ny)
					if !keep(cc) {
						seen[j] = 1
						continue
					}
					seen[j] = 2
					q = append(q, pt{nx, ny})
				}
			}
			if n > 80 {
				out = append(out, blob{minX, minY, maxX + 1, maxY + 1})
			}
		}
	}
	return out
}

func cropFeather(src *image.NRGBA, bl blob, pad int) *image.NRGBA {
	b := src.Bounds()
	x0 := max(b.Min.X, b.Min.X+bl.minX-pad)
	y0 := max(b.Min.Y, b.Min.Y+bl.minY-pad)
	x1 := min(b.Max.X, b.Min.X+bl.maxX+pad)
	y1 := min(b.Max.Y, b.Min.Y+bl.maxY+pad)
	r := image.Rect(0, 0, x1-x0, y1-y0)
	out := image.NewNRGBA(r)
	cx := float64(bl.minX+bl.maxX)/2 - float64(x0-b.Min.X)
	cy := float64(bl.minY+bl.maxY)/2 - float64(y0-b.Min.Y)
	rx := float64(bl.dx())/2 + float64(pad)
	ry := float64(bl.dy())/2 + float64(pad)
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			i := src.NRGBAAt(x, y)
			nx := (float64(x-x0) - cx) / rx
			ny := (float64(y-y0) - cy) / ry
			e := nx*nx + ny*ny
			var a float64 = 1
			if e > 0.82 {
				if e >= 1.05 {
					a = 0
				} else {
					a = 1 - (e-0.82)/(1.05-0.82)
				}
			}
			i.A = uint8(float64(255) * a)
			out.SetNRGBA(x-x0, y-y0, i)
		}
	}
	return out
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- coin.glb : disque PBR, textures face / pile ---

func buildCoinGLB(front, back image.Image) []byte {
	const nSeg = 64
	const radius = 1.0
	const thick = 0.09
	pos, nrm, uv, idx := coinMesh(nSeg, radius, thick)

	var geom bytes.Buffer
	writeF32 := func(v float32) { _ = binary.Write(&geom, binary.LittleEndian, v) }
	writeU16 := func(v uint16) { _ = binary.Write(&geom, binary.LittleEndian, v) }
	for _, v := range pos {
		writeF32(v)
	}
	posLen := geom.Len()
	for _, v := range nrm {
		writeF32(v)
	}
	nrmLen := geom.Len() - posLen
	for _, v := range uv {
		writeF32(v)
	}
	uvLen := geom.Len() - posLen - nrmLen
	for geom.Len()%4 != 0 {
		geom.WriteByte(0)
	}
	idxOff := geom.Len()
	for _, v := range idx {
		writeU16(v)
	}
	for geom.Len()%4 != 0 {
		geom.WriteByte(0)
	}
	imgOff := geom.Len()
	frontPNG := encodePNG(fitSquare(front, 512))
	geom.Write(frontPNG)
	frontLen := len(frontPNG)
	for geom.Len()%4 != 0 {
		geom.WriteByte(0)
	}
	backOff := geom.Len()
	backPNG := encodePNG(fitSquare(back, 512))
	geom.Write(backPNG)
	backLen := len(backPNG)
	for geom.Len()%4 != 0 {
		geom.WriteByte(0)
	}
	bin := geom.Bytes()

	minP, maxP := [3]float32{1e9, 1e9, 1e9}, [3]float32{-1e9, -1e9, -1e9}
	for i := 0; i < len(pos); i += 3 {
		for k := 0; k < 3; k++ {
			if pos[i+k] < minP[k] {
				minP[k] = pos[i+k]
			}
			if pos[i+k] > maxP[k] {
				maxP[k] = pos[i+k]
			}
		}
	}
	nVerts := len(pos) / 3
	json := fmt.Sprintf(`{
		"asset":{"version":"2.0","generator":"cd-scout-coin"},
		"scene":0,
		"scenes":[{"nodes":[0]}],
		"nodes":[{"mesh":0,"name":"Coin"}],
		"meshes":[{"name":"Coin","primitives":[{"attributes":{"POSITION":0,"NORMAL":1,"TEXCOORD_0":2},"indices":3,"material":0}]}],
		"materials":[{"name":"Gold","pbrMetallicRoughness":{"baseColorTexture":{"index":0},"metallicFactor":1.0,"roughnessFactor":0.28},"doubleSided":true}],
		"textures":[{"source":0},{"source":1}],
		"images":[
			{"mimeType":"image/png","bufferView":4,"name":"front"},
			{"mimeType":"image/png","bufferView":5,"name":"back"}
		],
		"samplers":[{"magFilter":9729,"minFilter":9987,"wrapS":10497,"wrapT":10497}],
		"accessors":[
			{"bufferView":0,"componentType":5126,"count":%d,"type":"VEC3","min":[%g,%g,%g],"max":[%g,%g,%g]},
			{"bufferView":1,"componentType":5126,"count":%d,"type":"VEC3"},
			{"bufferView":2,"componentType":5126,"count":%d,"type":"VEC2"},
			{"bufferView":3,"componentType":5123,"count":%d,"type":"SCALAR"}
		],
		"bufferViews":[
			{"buffer":0,"byteOffset":0,"byteLength":%d,"target":34962},
			{"buffer":0,"byteOffset":%d,"byteLength":%d,"target":34962},
			{"buffer":0,"byteOffset":%d,"byteLength":%d,"target":34962},
			{"buffer":0,"byteOffset":%d,"byteLength":%d,"target":34963},
			{"buffer":0,"byteOffset":%d,"byteLength":%d},
			{"buffer":0,"byteOffset":%d,"byteLength":%d}
		],
		"buffers":[{"byteLength":%d}]
	}`, nVerts,
		minP[0], minP[1], minP[2], maxP[0], maxP[1], maxP[2],
		nVerts, nVerts, len(idx),
		posLen, posLen, nrmLen, posLen+nrmLen, uvLen,
		idxOff, len(idx)*2,
		imgOff, frontLen, backOff, backLen,
		len(bin))
	json = strings.ReplaceAll(json, "\t", "")
	json = strings.ReplaceAll(json, "\n", "")
	return packGLB([]byte(json), bin)
}

func encodePNG(im image.Image) []byte {
	var b bytes.Buffer
	_ = png.Encode(&b, im)
	return b.Bytes()
}

func fitSquare(im image.Image, size int) *image.NRGBA {
	src := toNRGBA(im)
	out := image.NewNRGBA(image.Rect(0, 0, size, size))
	sb := src.Bounds()
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			sx := sb.Min.X + x*sb.Dx()/size
			sy := sb.Min.Y + y*sb.Dy()/size
			out.SetNRGBA(x, y, src.NRGBAAt(sx, sy))
		}
	}
	return out
}

func coinMesh(n int, r, t float32) (pos, nrm, uv []float32, idx []uint16) {
	// front center, front rim, back center, back rim, side (2 rings)
	zf, zb := t/2, -t/2
	add := func(x, y, z, nx, ny, nz, u, v float32) uint16 {
		i := uint16(len(pos) / 3)
		pos = append(pos, x, y, z)
		nrm = append(nrm, nx, ny, nz)
		uv = append(uv, u, v)
		return i
	}
	fc := add(0, 0, zf, 0, 0, 1, 0.5, 0.5)
	front := make([]uint16, n)
	for i := 0; i < n; i++ {
		a := float64(i) / float64(n) * 2 * math.Pi
		x, y := float32(math.Cos(a))*r, float32(math.Sin(a))*r
		u := 0.5 + float32(math.Cos(a))*0.5
		v := 0.5 + float32(math.Sin(a))*0.5
		front[i] = add(x, y, zf, 0, 0, 1, u, v)
	}
	bc := add(0, 0, zb, 0, 0, -1, 0.5, 0.5)
	back := make([]uint16, n)
	for i := 0; i < n; i++ {
		a := float64(i) / float64(n) * 2 * math.Pi
		x, y := float32(math.Cos(a))*r, float32(math.Sin(a))*r
		u := 0.5 - float32(math.Cos(a))*0.5
		v := 0.5 + float32(math.Sin(a))*0.5
		back[i] = add(x, y, zb, 0, 0, -1, u, v)
	}
	sideF := make([]uint16, n)
	sideB := make([]uint16, n)
	for i := 0; i < n; i++ {
		a := float64(i) / float64(n) * 2 * math.Pi
		nx, ny := float32(math.Cos(a)), float32(math.Sin(a))
		x, y := nx*r, ny*r
		u := float32(i) / float32(n)
		sideF[i] = add(x, y, zf, nx, ny, 0, u, 0)
		sideB[i] = add(x, y, zb, nx, ny, 0, u, 1)
	}
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		idx = append(idx, fc, front[i], front[j])
		idx = append(idx, bc, back[j], back[i])
		idx = append(idx, sideF[i], sideB[i], sideF[j])
		idx = append(idx, sideF[j], sideB[i], sideB[j])
	}
	return
}

func packGLB(js, bin []byte) []byte {
	for len(js)%4 != 0 {
		js = append(js, ' ')
	}
	for len(bin)%4 != 0 {
		bin = append(bin, 0)
	}
	total := 12 + 8 + len(js) + 8 + len(bin)
	out := make([]byte, 0, total)
	putU32 := func(v uint32) {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], v)
		out = append(out, b[:]...)
	}
	out = append(out, 'g', 'l', 'T', 'F')
	putU32(2)
	putU32(uint32(total))
	putU32(uint32(len(js)))
	putU32(0x4E4F534A) // JSON
	out = append(out, js...)
	putU32(uint32(len(bin)))
	putU32(0x004E4942) // BIN
	out = append(out, bin...)
	return out
}
