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
    "code.google.com/p/go.image/bmp"
)

func main(){
    var input_name = make([]string, 2)
    var output_name string
    var fp_src = make([]*os.File, 2)
    var fp_dst *os.File
    var img_src = make([]image.Image, 2)
    var err error
    var argv int = len(os.Args)
    var p_img_dst *image.RGBA
    var gain uint32 = 1
    var rgb_mode string = "rgb"
    var input_ext = make([]string, 2)

    var p_flag_out = flag.String("o", "out.jpg", "-o OUTPUT_FILE_NAME")
    var p_flag_rgb = flag.String("c", "rgb", "-c RGB_TYPE (Range [rgb, r, g, b])")
    var p_flag_gain = flag.Int("g", 1, "-g GAIN (Range [1, 10])")

    flag.Parse()
    fmt.Println(*p_flag_rgb)
    output_name = *p_flag_out
    gain = (uint32)(*p_flag_gain)
    rgb_mode = *p_flag_rgb

    if(argv < 3){
        fmt.Println("Usage: image_comp (optins..) INPUT_FILE01 INPUT_FILE02")
        fmt.Println("       [Options]")
        fmt.Println("           -o OUTPUT_FILE_NAME (without file extension) (Default: out.bmp)")
        fmt.Println("           -c PIXEL_TO_COMPARE (Range: [rgb, r, g, b] Default: rgb)")
        fmt.Println("           -g GAIN (Range [1, 10], Default: 1)")
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
    fmt.Println("color mode : ", rgb_mode);
    fmt.Println("gain       : ", gain)

    /* Read input images */
    for i := 0; i < 2; i++ {
        fp_src[i], err = os.Open(input_name[i]);
        if err != nil{
            fmt.Println("Error: ", err);
        }
        defer fp_src[i].Close()

        switch input_ext[0] {
        case ".bmp":
            img_src[i], err = bmp.Decode(fp_src[i])
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

    /* Make diff image */
    p_img_dst = makeDiff(img_src[0], img_src[1], rgb_mode, gain)

    /* Write diff image */
    {
        fp_dst, err = os.Create(output_name);
        if err != nil{
            fmt.Println("Error: ", err);
        }
        defer fp_dst.Close()

        switch path.Ext(output_name) {
        case ".bmp":
            err = bmp.Encode(fp_dst, p_img_dst)
        case ".jpg":
            err = jpeg.Encode(fp_dst, p_img_dst, &jpeg.Options{100})
        default:
            fmt.Println("Error: Unsupported output image file format.")
            return
        }

        if err != nil{
            fmt.Println("Error: ", err);
        }
    }

    fmt.Println("\nSuccess!!");
}

func makeDiff(img_base, img_comp image.Image, rgb_mode string, gain uint32) *image.RGBA{
    var p_img_diff *image.RGBA
    var r0, g0, b0, r1, g1, b1 uint32
    var pr0, pg0, pb0, pr1, pg1, pb1 *uint32
    var sx, sy = 0, 0
    var ex, ey = img_base.Bounds().Max.X, img_base.Bounds().Max.Y

    switch rgb_mode {
    case "rgb":
        pr0, pg0, pb0 = &r0, &g0, &b0
        pr1, pg1, pb1 = &r1, &g1, &b1
    case "r":
        pr0, pg0, pb0 = &r0, &r0, &r0
        pr1, pg1, pb1 = &r1, &r1, &r1
    case "g":
        pr0, pg0, pb0 = &g0, &g0, &g0
        pr1, pg1, pb1 = &g1, &g1, &g1
    case "b":
        pr0, pg0, pb0 = &b0, &b0, &b0
        pr1, pg1, pb1 = &b1, &b1, &b1
    default:
        /* Do Nothing.. */
    }

    p_img_diff = image.NewRGBA( image.Rect(sx, sy, ex, ey) )
    for y := sx; y < ey; y++ {
        for x := sy; x < ex; x++ {
            r0, g0, b0, _ = img_base.At(x, y).RGBA()
            r1, g1, b1, _ = img_comp.At(x, y).RGBA()

            rr := float64( ( 128<<8 + (*pr0 - *pr1) * (gain) + 1<<7) >> 8 )
            gg := float64( ( 128<<8 + (*pg0 - *pg1) * (gain) + 1<<7) >> 8 )
            bb := float64( ( 128<<8 + (*pb0 - *pb1) * (gain) + 1<<7) >> 8 )

            rr = math.Min( math.Max(0.0, rr), 255.0 )
            gg = math.Min( math.Max(0.0, gg), 255.0 )
            bb = math.Min( math.Max(0.0, bb), 255.0 )

            p_img_diff.SetRGBA(x, y, color.RGBA{uint8(rr), uint8(gg), uint8(bb), 0xff})
        }
    }
    return p_img_diff
}
