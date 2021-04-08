video2frames is a utility to convert video files into single image frames. Depends on ffmpeg CLI utilities.

## Main Features
- Uncompressed BMP or Compressed JPEG output
- Grayscale output
- Set output dimensions
- Set output factor (e.g. skip every other frame: `video2frames -x 50...`)
- write custom exifdata for use with popular SFM and photogrammetry software

## Dependancies
- exiftool
- ffmpeg

## Example use case:

`video2frames -i source.mp4 -o ./exported_frames`
