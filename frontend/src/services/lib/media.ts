
export class MediaController {
    stream: MediaStream | null = null

    async getAV(): Promise<MediaStream> {
        if (!this.stream){
            this.stream =await navigator.mediaDevices.getUserMedia({
            video: {
                width: { ideal: 320, max: 320 },
                height: { ideal: 180, max: 180 },
                frameRate: { ideal: 15, max: 30 },
            },
            audio: {
                channelCount: 1,
                sampleRate: 16000,
                echoCancellation: true,
                noiseSuppression: true,
                autoGainControl: true,
            },
            });

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