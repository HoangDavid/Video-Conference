
export class MediaController {
    stream: MediaStream | null = null

    async getAV(): Promise<MediaStream> {
        if (!this.stream){
            this.stream = await navigator.mediaDevices.getUserMedia({video: true, audio: true});
        }

        return this.stream;
    }

    get audio() {return this.stream?.getAudioTracks()[0] ?? null;}
    get video() {return this.stream?.getVideoTracks()[0] ?? null;}

    set_Mic(flag: boolean) {const a = this.audio; if (a) a.enabled = flag;} 
    set_Video(flag: boolean) {const v = this.video; if (v) v.enabled = flag;}

    remove() {this.stream?.getTracks().forEach(t => t.stop()); this.stream = null;}
}

export const media = new MediaController();