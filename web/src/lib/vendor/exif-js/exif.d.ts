interface EXIFStatic {
    getData(url: { src: string } | HTMLImageElement | File, callback: (photo: any) => void): any;
    getTag(img: { src: string } | HTMLImageElement | File, tag: any): any;
    getAllTags(img: string | File): any;
    pretty(img: any): string;
    readFromBinaryFile(file: any): any;
}

declare var EXIF: EXIFStatic;
export = EXIF;