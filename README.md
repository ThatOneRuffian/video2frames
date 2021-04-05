# video_to_frames
video2frames is a utility to convert video files into single image frames. Depends on ffmpeg CLI utilities.

## Main Features
- Uncompressed BMP or Compressed PNG output
- Grayscale output
- Set output dimensions
- Set output factor (e.g. skip every other frame: `video2frames -x 50...`)

## Example use case:

`video2frames -i pipe.mp4 -o ./exported_frames`
