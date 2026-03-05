* [x] Fix reisen build in CI and display bug (fixed with build tags and corrected timing).
* [x] Re-implement VIDEO stimuli using the "Extreme Gpu Friendly Video Format" (.gv) - FIXED (Pure Go, No CGo). Users need to convert movies to .gv format using Ushio's converter.

The video solution is not satisfaying, the reisen library is deprecated, and cgo is painful. We need to restart from scrath the implemenation of the VIDEO stimuli. I suggest to use of the "Extreme Gpu Friendly Video Format" which can be played with pure go (see ebiten_gvvideo). sequences of images can be saved in this format (Converters from mov or mp4 exist https://github.com/Ushio/ofxExtremeGpuVideo). This users will need to convert by themselves the video file into .gv format, but this is fine.
