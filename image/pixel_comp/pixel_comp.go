package main

import (
    "fmt"
    "os"
    "flag"
    "path"
    "image"
    "image/jpeg"
    "image/color"
    "math"
    //"image/bmp"
)

func main(){
    var input_name = make([]string, 2)
    var output_name string
    var fp_src = make([]*os.File, 2)
    var fp_dst *os.File
    var img_src = make([]image.Image, 2)
    var err error
    var argv int = len(os.Args)
    var input_ext = make([]string, 2)

    var p_flag_out = flag.String("o", "out.jpg", "-o OUTPUT_FILE_NAME")

    flag.Parse()
    output_name = *p_flag_out

    if(argv < 3){
        fmt.Println("Usage: image_comp (optins..) INPUT_FILE01 INPUT_FILE02")
        fmt.Println("       [Options]")
        fmt.Println("           -o OUTPUT_FILE_NAME (without file extension) (Default: out.bmp)")
        fmt.Println("           -c PIXEL_TO_COMPARE (Range: [rgb, r, g, b] Default: rgb)")
        fmt.Println("Example: image_comp -o result.bmp -g 2 input01.jpg input02.jpg")
        fmt.Println("Note: Supported input/output file types are BMP and JPG.")
        fmt.Println("      If OUTPUT_FILE_NAME is 'xxx.bmp' ('xxx.jpg'), output file format is BMP (JPG).")
        return
    }

    input_name[0] = os.Args[argv-2]
    input_name[1] = os.Args[argv-1]

    input_ext[0] = path.Ext(input_name[0])
    input_ext[1] = path.Ext(input_name[1])
    if input_ext[0] != input_ext[1] {
        fmt.Println("Error: Input image file types are different.")
        return
    }

    fmt.Println("input1     : ", input_name[0]);
    fmt.Println("input2     : ", input_name[1]);
    fmt.Println("output     : ", output_name);

    /* Read input images */
    for i := 0; i < 2; i++ {
        fp_src[i], err = os.Open(input_name[i]);
        if err != nil{
            fmt.Println("Error: ", err);
        }
        defer fp_src[i].Close()

        switch input_ext[0] {
//        case ".bmp":
//            img_src[i], err = bmp.Decode(fp_src[i])
        case ".jpg":
            img_src[i], _, err = image.Decode(fp_src[i])
        default:
            fmt.Println("Error: Unsupported input image file format.")
            return
        }

        if err != nil{
            fmt.Println("Error: ", err);
        }
    }

    file_out, err := os.Create("test.txt")
    if err != nil{
        fmt.Println("Error: ", err)
    }

    var sx, sy = 0, 0
    var ex, ey = img_src[0].Bounds().Max.X, img_src[0].Bounds().Max.Y

    var p_img_conv *image.RGBA
    p_img_conv = image.NewRGBA( image.Rect(sx, sy, ex, ey) )

    for y := sx; y < ey; y++ {
        for x := sy; x < ex; x++ {
            var r0, g0, b0 uint32
            var r1, g1, b1 uint32
            var rr0, gg0, bb0 uint8
            var h0, s0, v0 uint32
            var h1 uint32

            r0, g0, b0, _ = img_src[0].At(x, y).RGBA()
            r1, g1, b1, _ = img_src[1].At(x, y).RGBA()
            h0, s0, v0 = rgb2hsv(r0, g0, b0)
            h1, _, _ = rgb2hsv(r1, g1, b1)

            file_out.Write( []byte(fmt.Sprint(h0)) )
            file_out.Write( []byte(", ") )
            file_out.Write( []byte(fmt.Sprint(h1)) )
            file_out.Write( []byte("\n") )
            rr0, gg0, bb0 = hsv2rgb( h0, s0, v0 )

            p_img_conv.SetRGBA(x, y, color.RGBA{rr0, gg0, bb0, 0xff})
        }
    }

    {
        fp_dst, err = os.Create(output_name)
        if err != nil{
            fmt.Println("Error: ", err);
        }
        defer fp_dst.Close()

        err = jpeg.Encode(fp_dst, p_img_conv, &jpeg.Options{100})
        if err != nil{
            fmt.Println("Error: ", err)
        }
    }

    fmt.Println("\nSuccess!!");
}

/*
    r: Range[0, 255]
    g: Range[0, 255]
    b: Range[0, 255]

    h: Range[0, 360]
    s: Range[0, 255]
    v: Range[0, 255]
*/
func rgb2hsv(r, g, b uint32) (hh, ss, vv uint32) {
    var h int32
    var s, v int32

    fR, fG, fB := int32(r>>8), int32(g>>8), int32(b>>8)

    max := int32( math.Max( math.Max( float64(fR), float64(fG) ), float64(fB) ) )
    min := int32( math.Min( math.Min( float64(fR), float64(fG) ), float64(fB) ) )
    d := max - min
    s, v = 0, max
    if max > 0 {
        s = (d*255) / max
    }
    if max == min {
        h = 0
    } else {
        switch max {
        case fR:
            h = ((fG - fB)<<8) / d
            if fG < fB {
                    h += (6<<8)
            }
        case fG:
            h = ((fB-fR)<<8)/d + (2<<8)
        case fB:
            h = ((fR-fG)<<8)/d + (4<<8)
        }

        h *= 60
        h = (h>>8)
    }

    hh, ss, vv = uint32(h), uint32(s), uint32(v)

    return
}

func hsv2rgb(hh, ss, vv uint32) (r, g, b uint8) {
    var h, s, v int32
    var fR, fG, fB int32
    var i, f, p, q, t int32

    h, s, v = int32(hh), int32(ss), int32(vv)
    h = ((h<<8)/60)
    i = (h>>8)
    f = h - (i<<8)

    p = (v * (255 - s)) >> 8
    q = (v * ((255<<8) - f*s)) >> 16
    t = (v * ((255<<8) - (255-f)*s)) >> 16

    switch i {
    case 0:
            fR, fG, fB = v, t, p
    case 1:
            fR, fG, fB = q, v, p
    case 2:
            fR, fG, fB = p, v, t
    case 3:
            fR, fG, fB = p, q, v
    case 4:
            fR, fG, fB = t, p, v
    case 5:
            fR, fG, fB = v, p, q
    }

    r = uint8(fR)
    g = uint8(fG)
    b = uint8(fB)

    return
}
