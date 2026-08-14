# Compose les bannières d'alerte : art 256 px + titres/tips FR (PNG contour) + bandeau.
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$artDir = Join-Path $here 'art'
$textDir = Join-Path $here 'text'
New-Item -ItemType Directory -Force -Path $artDir, $textDir | Out-Null

$E = [char]0x00C9
$e = [char]0x00E9
$A = [char]0x00C2
$egr = [char]0x00E8

$titleFam = New-Object System.Drawing.Text.PrivateFontCollection
$titleFam.AddFontFile('C:\Windows\Fonts\arialbd.ttf')
$tipFam = New-Object System.Drawing.Text.PrivateFontCollection
$tipFam.AddFontFile('C:\Windows\Fonts\arialbd.ttf')
$familyTitle = $titleFam.Families[0]
$familyTip = $tipFam.Families[0]

function Save-Png([System.Drawing.Bitmap]$bmp, [string]$path) {
	$tmp = "$path.tmp"
	$bmp.Save($tmp, [System.Drawing.Imaging.ImageFormat]::Png)
	if (Test-Path $path) { Remove-Item $path -Force }
	Move-Item $tmp $path -Force
}

function Resize-Art([string]$path, [int]$size) {
	$src = [System.Drawing.Image]::FromFile($path)
	if ($src.Width -eq $size -and $src.Height -eq $size) { $src.Dispose(); return }
	$bmp = New-Object System.Drawing.Bitmap $size, $size
	$g = [System.Drawing.Graphics]::FromImage($bmp)
	$g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
	$g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
	$g.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
	$g.DrawImage($src, 0, 0, $size, $size)
	$src.Dispose(); $g.Dispose()
	Save-Png $bmp $path
	$bmp.Dispose()
}

function Add-Round([System.Drawing.Drawing2D.GraphicsPath]$p, [float]$x, [float]$y, [float]$w, [float]$h, [float]$r) {
	$d = $r * 2
	$p.AddArc($x, $y, $d, $d, 180, 90)
	$p.AddArc($x + $w - $d, $y, $d, $d, 270, 90)
	$p.AddArc($x + $w - $d, $y + $h - $d, $d, $d, 0, 90)
	$p.AddArc($x, $y + $h - $d, $d, $d, 90, 90)
	$p.CloseFigure()
}

function Draw-Outline([System.Drawing.Graphics]$g, [string]$text, $family, [float]$em, [float]$x, [float]$y, $fill, [float]$stroke, [int]$style) {
	$gp = New-Object System.Drawing.Drawing2D.GraphicsPath
	$sf = New-Object System.Drawing.StringFormat
	$sf.Alignment = [System.Drawing.StringAlignment]::Near
	$sf.LineAlignment = [System.Drawing.StringAlignment]::Near
	$sf.FormatFlags = [System.Drawing.StringFormatFlags]::NoWrap
	$gp.AddString($text, $family, $style, $em, (New-Object System.Drawing.PointF $x, $y), $sf)
	$pen = New-Object System.Drawing.Pen ([System.Drawing.Color]::FromArgb(255, 17, 17, 17), $stroke)
	$pen.LineJoin = [System.Drawing.Drawing2D.LineJoin]::Round
	$g.DrawPath($pen, $gp)
	$br = New-Object System.Drawing.SolidBrush $fill
	$g.FillPath($br, $gp)
	$b = $gp.GetBounds()
	$gp.Dispose(); $pen.Dispose(); $br.Dispose(); $sf.Dispose()
	return $b
}

function Fit-Outline([System.Drawing.Graphics]$g, [string]$text, $family, [float]$em, [float]$minEm, [float]$x, [float]$y, [float]$maxW, $fill, [float]$stroke, [int]$style) {
	$dummy = New-Object System.Drawing.Bitmap 8, 8
	$dg = [System.Drawing.Graphics]::FromImage($dummy)
	$dg.PageUnit = [System.Drawing.GraphicsUnit]::Pixel
	$size = $em
	while ($size -gt $minEm) {
		$gp = New-Object System.Drawing.Drawing2D.GraphicsPath
		$sf = New-Object System.Drawing.StringFormat
		$sf.FormatFlags = [System.Drawing.StringFormatFlags]::NoWrap
		$gp.AddString($text, $family, $style, $size, (New-Object System.Drawing.PointF 0, 0), $sf)
		$w = $gp.GetBounds().Width
		$gp.Dispose(); $sf.Dispose()
		if ($w -le $maxW) { break }
		$size -= 1.5
	}
	$dg.Dispose(); $dummy.Dispose()
	return Draw-Outline $g $text $family $size $x $y $fill $stroke $style
}

function Measure-Width([string]$text, $family, [float]$em, [int]$style) {
	$gp = New-Object System.Drawing.Drawing2D.GraphicsPath
	$sf = New-Object System.Drawing.StringFormat
	$sf.FormatFlags = [System.Drawing.StringFormatFlags]::NoWrap
	$gp.AddString($text, $family, $style, $em, (New-Object System.Drawing.PointF 0, 0), $sf)
	$w = $gp.GetBounds().Width
	$gp.Dispose(); $sf.Dispose()
	return $w
}

function Split-Tip([string]$text, $family, [float]$em, [float]$maxW, [int]$style) {
	$colon = $text.LastIndexOf(' : ')
	if ($colon -lt 0) { $colon = $text.LastIndexOf(': ') }
	if ($colon -gt 0) {
		$a = $text.Substring(0, $colon).TrimEnd(' ', ':')
		$b2 = $text.Substring($colon).TrimStart(' ', ':').Trim()
		return @($a, $b2)
	}
	$comma = $text.LastIndexOf(',')
	if ($comma -gt 0) {
		$a = $text.Substring(0, $comma + 1).Trim()
		$b2 = $text.Substring($comma + 1).Trim()
		if ($b2) { return @($a, $b2) }
	}
	$words = $text.Split(' ')
	if ($words.Length -le 3) { return @($text) }
	$mid = [Math]::Max(1, [Math]::Floor($words.Length / 2))
	while ($mid -gt 1 -and (Measure-Width ($words[0..($mid - 1)] -join ' ') $family $em $style) -gt $maxW) { $mid-- }
	return @(($words[0..($mid - 1)] -join ' '), ($words[$mid..($words.Length - 1)] -join ' '))
}

function Draw-Tip([System.Drawing.Graphics]$g, [string]$text, $family, [float]$em, [float]$x, [float]$y, [float]$maxW, $fill, [float]$stroke, [int]$style) {
	$lines = @(Split-Tip $text $family $em $maxW $style)
	$yy = $y
	foreach ($line in $lines) {
		[void](Draw-Outline $g $line $family $em $x $yy $fill $stroke $style)
		$yy += $em + 8
	}
}

function New-TextPng([string]$text, [string]$path, [int]$w, [int]$h, $family, [float]$em, $fill, [float]$stroke, [int]$style) {
	$bmp = New-Object System.Drawing.Bitmap $w, $h, ([System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
	$g = [System.Drawing.Graphics]::FromImage($bmp)
	$g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
	$g.PageUnit = [System.Drawing.GraphicsUnit]::Pixel
	$g.Clear([System.Drawing.Color]::FromArgb(255, 255, 0, 255))
	[void](Draw-Outline $g $text $family $em 18 10 $fill $stroke $style)
	$g.Dispose()
	$bmp.MakeTransparent([System.Drawing.Color]::FromArgb(255, 255, 0, 255))
	Save-Png $bmp $path
	$bmp.Dispose()
}

Get-ChildItem $artDir -Filter '*.png' | ForEach-Object { Resize-Art $_.FullName 256 }

$white = [System.Drawing.Color]::FromArgb(255, 255, 254, 244)
$styleReg = [int][System.Drawing.FontStyle]::Regular
$styleBold = [int][System.Drawing.FontStyle]::Bold

$banners = @(
	@{ id = 'grubs';     art = 'grubs';     bg = '#d4c4ff'; title = "LARVES DANS 1:15";                         tip = "Pousse ta vague et monte au pit" },
	@{ id = 'herald';    art = 'herald';    bg = '#ffa3d4'; title = "H${E}RAUT DANS 1:15";                       tip = "Ward le pit, pr${e}pare le si${egr}ge" },
	@{ id = 'baron';     art = 'baron';     bg = '#c4b0ff'; title = "BARON DANS 1:15";                          tip = "Vision mid + jungle, regroupe-toi" },
	@{ id = 'dragon';    art = 'dragon';    bg = '#ffb26b'; title = "DRAKE DANS 1:15";                          tip = "Prends la vision bot et regroupe-toi" },
	@{ id = 'infernal';  art = 'infernal';  bg = '#ff9a4a'; title = "DRAKE INFERNAL DANS 1:15";                  tip = "Fight burst : vision bot, groupe" },
	@{ id = 'mountain';  art = 'mountain';  bg = '#e0c48a'; title = "DRAKE MONTAGNE DANS 1:15";                  tip = "Fight tanky : vision bot, groupe" },
	@{ id = 'ocean';     art = 'ocean';     bg = '#7ec8e8'; title = "DRAKE OC${E}AN DANS 1:15";                   tip = "Fight sustain : vision bot, groupe" },
	@{ id = 'cloud';     art = 'cloud';     bg = '#b8dcff'; title = "DRAKE NUAGE DANS 1:15";                     tip = "Fight speed : ward et groupe" },
	@{ id = 'hextech';   art = 'hextech';   bg = '#7ad4e8'; title = "DRAKE HEXTECH DANS 1:15";                   tip = "Hexgates : vision bot et portes" },
	@{ id = 'chemtech';  art = 'chemtech';  bg = '#b8e07a'; title = "DRAKE CHEMTECH DANS 1:15";                  tip = "Plantes mutantes : vision bot, groupe" },
	@{ id = 'soul';      art = 'infernal';  bg = '#ffd76b'; title = "DRAKE D'${A}ME DANS 1:15";                   tip = "Teamfight d${e}cisif : vision bot maintenant" },
	@{ id = 'elder';     art = 'elder';     bg = '#e8d48a'; title = "ANCESTRAL DANS 1:15";                       tip = "Le fight qui ferme la partie" }
)

function Parse-Hex([string]$h) {
	$h = $h.TrimStart('#')
	return [System.Drawing.Color]::FromArgb(255, [Convert]::ToInt32($h.Substring(0, 2), 16), [Convert]::ToInt32($h.Substring(2, 2), 16), [Convert]::ToInt32($h.Substring(4, 2), 16))
}

$W = 740; $H = 200
foreach ($b in $banners) {
	New-TextPng $b.title (Join-Path $textDir "$($b.id)-title.png") 640 72 $familyTitle 36 $white 7 $styleBold
	$tipPng = New-Object System.Drawing.Bitmap 640, 100, ([System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
	$tg = [System.Drawing.Graphics]::FromImage($tipPng)
	$tg.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
	$tg.PageUnit = [System.Drawing.GraphicsUnit]::Pixel
	$tg.Clear([System.Drawing.Color]::FromArgb(255, 255, 0, 255))
	Draw-Tip $tg $b.tip $familyTip 30 18 8 600 $white 6 $styleBold
	$tg.Dispose()
	$tipPng.MakeTransparent([System.Drawing.Color]::FromArgb(255, 255, 0, 255))
	Save-Png $tipPng (Join-Path $textDir "$($b.id)-tip.png")
	$tipPng.Dispose()

	$bg = Parse-Hex $b.bg
	$bmp = New-Object System.Drawing.Bitmap $W, $H, ([System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
	$g = [System.Drawing.Graphics]::FromImage($bmp)
	$g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
	$g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
	$g.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
	$g.PageUnit = [System.Drawing.GraphicsUnit]::Pixel
	$g.Clear([System.Drawing.Color]::FromArgb(255, 17, 17, 17))

	$inner = New-Object System.Drawing.Drawing2D.GraphicsPath
	Add-Round $inner 5 5 ($W - 10) ($H - 10) 20
	$g.FillPath((New-Object System.Drawing.SolidBrush $bg), $inner)
	$inner.Dispose()

	$artPath = Join-Path $artDir "$($b.art).png"
	$art = [System.Drawing.Image]::FromFile($artPath)
	$ax = 14; $ay = 16; $as = 168
	$frame = New-Object System.Drawing.Drawing2D.GraphicsPath
	Add-Round $frame $ax $ay $as $as 18
	$g.FillPath((New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(255, 17, 17, 17))), $frame)
	$g.SetClip($frame)
	$g.DrawImage($art, $ax + 4, $ay + 4, $as - 8, $as - 8)
	$g.ResetClip()
	$frame.Dispose()
	$art.Dispose()

	$tx = 196
	[void](Fit-Outline $g $b.title $familyTitle 34 18 $tx 22 520 $white 6.5 $styleBold)
	Draw-Tip $g $b.tip $familyTip 28 $tx 88 520 $white 6 $styleBold

	$g.Dispose()
	Save-Png $bmp (Join-Path $here "$($b.id).png")
	$bmp.Dispose()
	Write-Host "ok $($b.id)"
}

Write-Host 'done'
