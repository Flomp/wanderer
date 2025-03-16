import type PhotoSwipe from "photoswipe";
import type PhotoSwipeLightbox from "photoswipe/lightbox";
import type { Content } from "photoswipe/lightbox";
import type { SlideData } from "photoswipe";
import type { PhotoSwipeVideoPluginOptions } from "./options";

type VideoContentData = {
    width: number,
    height: number,
    type: 'video',
    msrc: string,
    videoSrc: string,
}

type VideoContent = Content & {
    element: HTMLVideoElement
    data: VideoContentData
    _videoPosterImg: HTMLImageElement | null
}

function isVideoContent(content?: Content | SlideData) {
    if (!content) {
        return false
    }
    return (content && content.data && content.data.type === 'video');
}

class VideoContentSetup {
    options: PhotoSwipeVideoPluginOptions;
    constructor(lightbox: PhotoSwipeLightbox, options: PhotoSwipeVideoPluginOptions) {
        this.options = options;

        this.initLightboxEvents(lightbox);
        lightbox.on('init', () => {
            this.initPswpEvents(lightbox.pswp!);
        });
    }

    initLightboxEvents(lightbox: PhotoSwipeLightbox) {
        lightbox.on('contentLoad', this.onContentLoad.bind(this));
        lightbox.on('contentDestroy', this.onContentDestroy.bind(this));
        lightbox.on('contentActivate', this.onContentActivate.bind(this));
        lightbox.on('contentDeactivate', this.onContentDeactivate.bind(this));
        lightbox.on('contentAppend', this.onContentAppend.bind(this));
        lightbox.on('contentResize', this.onContentResize.bind(this));

        lightbox.addFilter('isKeepingPlaceholder', this.isKeepingPlaceholder.bind(this));
        lightbox.addFilter('isContentZoomable', this.isContentZoomable.bind(this));
        lightbox.addFilter('useContentPlaceholder', this.useContentPlaceholder.bind(this));

        lightbox.addFilter('domItemData', (itemData, element, linkEl) => {
            if (itemData.type === 'video' && linkEl) {
                if (linkEl.dataset.pswpVideoSources) {
                    itemData.videoSources = JSON.parse(linkEl.dataset.pswpVideoSources);
                } else if (linkEl.dataset.pswpVideoSrc) {
                    itemData.videoSrc = linkEl.dataset.pswpVideoSrc;
                } else {
                    itemData.videoSrc = linkEl.href;
                }
            }
            return itemData;
        });
    }

    initPswpEvents(pswp: PhotoSwipe) {
        pswp.on('pointerDown', (e) => {
            const slide = pswp.currSlide;
            if (isVideoContent(slide) && this.options.preventDragOffset) {
                const origEvent = e.originalEvent;
                if (origEvent.type === 'pointerdown') {
                    const videoHeight = Math.ceil((slide?.height ?? 0) * (slide?.currZoomLevel ?? 0));
                    const verticalEnding = videoHeight + (slide?.bounds.center.y ?? 0);
                    const pointerYPos = origEvent.pageY - pswp.offset.y;
                    if (pointerYPos > verticalEnding - this.options.preventDragOffset
                        && pointerYPos < verticalEnding) {
                        e.preventDefault();
                    }
                }
            }
        });

        // do not append video on nearby slides
        pswp.on('appendHeavy', (e) => {
            if (isVideoContent(e.slide) && !e.slide.isActive) {
                e.preventDefault();
            }
        });

        pswp.on('close', () => {
            if (isVideoContent(pswp.currSlide?.content)) {
                // Switch from zoom to fade closing transition,
                // as zoom transition is choppy for videos
                if (!pswp.options.showHideAnimationType
                    || pswp.options.showHideAnimationType === 'zoom') {
                    pswp.options.showHideAnimationType = 'fade';
                }

                // pause video when closing
                this.pauseVideo(pswp.currSlide?.content as VideoContent);
            }
        });
    }

    onContentDestroy({ content }: { content: Content }) {
        if (isVideoContent(content)) {
            const c = content as VideoContent
            if (c._videoPosterImg) {
                c._videoPosterImg.onload = c._videoPosterImg.onerror = null;
                c._videoPosterImg = null;
            }
        }
    }

    onContentResize(e: any) {
        if (isVideoContent(e.content)) {
            e.preventDefault();

            const width = e.width;
            const height = e.height;
            const content = e.content;

            if (content.element) {
                content.element.style.width = width + 'px';
                content.element.style.height = height + 'px';
            }

            if (content.slide && content.slide.placeholder) {
                // override placeholder size, so it more accurately matches the video
                const placeholderElStyle = content.slide.placeholder.element.style;
                placeholderElStyle.transform = 'none';
                placeholderElStyle.width = width + 'px';
                placeholderElStyle.height = height + 'px';
            }
        }
    }


    isKeepingPlaceholder(isZoomable: boolean, content: Content) {
        if (isVideoContent(content)) {
            return false;
        }
        return isZoomable;
    }

    isContentZoomable(isZoomable: boolean, content: Content) {
        if (isVideoContent(content)) {
            return false;
        }
        return isZoomable;
    }

    onContentActivate({ content }: { content: Content }) {
        if (isVideoContent(content) && this.options.autoplay) {
            this.playVideo(content as VideoContent);
        }
    }

    onContentDeactivate({ content }: { content: Content }) {
        if (isVideoContent(content)) {
            this.pauseVideo(content as VideoContent);
        }
    }

    onContentAppend(e: any) {
        if (isVideoContent(e.content)) {
            e.preventDefault();
            e.content.isAttached = true;
            e.content.appendImage();
        }
    }

    onContentLoad(e: any) {
        const content = e.content; // todo: videocontent

        if (!isVideoContent(e.content)) {
            return;
        }

        // stop default content load
        e.preventDefault();

        if (content.element) {
            return;
        }

        content.state = 'loading';
        content.type = 'video'; // TODO: move this to pswp core?

        content.element = document.createElement('video');

        if (this.options.videoAttributes) {
            for (let key in this.options.videoAttributes) {
                content.element.setAttribute(key, this.options.videoAttributes[key as keyof PhotoSwipeVideoPluginOptions['videoAttributes']] || '');
            }
        }

        content.onLoaded();
        // content.element.setAttribute('poster', content.data.msrc);

        // this.preloadVideoPoster(content, content.data.msrc);

        // content.element.style.position = 'absolute';
        content.element.style.left = "0";
        content.element.style.top = "0";

        if (content.data.videoSrc) {
            content.element.src = content.data.videoSrc;
        }
    }

    preloadVideoPoster(content: VideoContent, src: string) {
        if (!content._videoPosterImg && src) {
            content._videoPosterImg = new Image();
            content._videoPosterImg.src = src;
            if (content._videoPosterImg.complete) {
                content.onLoaded();
            } else {
                content._videoPosterImg.onload = content._videoPosterImg.onerror = () => {
                    content.onLoaded();
                };
            }
        }
    }


    playVideo(content: VideoContent) {
        if (content.element) {
            content.element.play();
        }
    }

    pauseVideo(content?: VideoContent) {
        if (!content) {
            return;
        }
        if (content.element) {
            content.element.pause();
        }
    }

    useContentPlaceholder(usePlaceholder: boolean, content: Content) {
        if (isVideoContent(content)) {
            return true;
        }
        return usePlaceholder;
    }

}

export default VideoContentSetup;