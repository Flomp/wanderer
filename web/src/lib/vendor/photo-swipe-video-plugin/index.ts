import type PhotoSwipeLightbox from 'photoswipe/lightbox';
import { defaultOptions, type PhotoSwipeVideoPluginOptions } from './options';
import VideoContentSetup from './video-content-setup';

class PhotoSwipeVideoPlugin {
  constructor(lightbox: PhotoSwipeLightbox, options?: PhotoSwipeVideoPluginOptions) {
    new VideoContentSetup(lightbox, {
      ...defaultOptions,
      ...options
    });
  }
}

export default PhotoSwipeVideoPlugin;