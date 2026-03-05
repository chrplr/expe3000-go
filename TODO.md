* [x] Fix reisen build in CI and display bug (fixed with build tags and corrected timing).
* The video solution is not satisfaying, the reisen library is deprecated, and cgo is painful. We need to investigate the use of the "Extreme Gpu Friendly Video Format"
 which can be played with pure go (see ebiten_gvvideo). sequence of images can be save in this format. Converters from mo or mp4 exist https://github.com/Ushio/ofxExtremeGpuVideo.
