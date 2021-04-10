video2frames is a utility to convert video files into single image frames. Also supports writing of common photogrammetry exif tags.

## Main Features
- Uncompressed BMP or Compressed JPEG output
- Grayscale output
- Set output dimensions
- Set output factor (e.g. skip every other frame: `video2frames -x 50...`)
- Set output quality
- Write custom exif data for use with popular SFM and photogrammetry software (JPEG only)
- Dump exif data for images

## Dependancies
- exiftool
- ffmpeg

## Example use case:

`video2frames -i source.mp4 -o ./exported_frames`

## Typical workflow:

`video2frames --export-exif-template`  // Fill out the generated JSON file with desired exif data

`video2frames -x 30 -c --exif-data ./exif_data.JSON -i source.mp4 -o ./exported_frames`
