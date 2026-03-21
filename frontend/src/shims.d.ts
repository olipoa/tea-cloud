declare module "@worstone/vue-aplayer" {
  import { DefineComponent } from "vue";
  const APlayer: DefineComponent<{
    audio?: { name?: string; url: string; artist?: string; cover?: string }[];
    autoplay?: boolean;
    loop?: "one" | "all" | "none";
    order?: "list" | "random";
    theme?: string;
    listFolded?: boolean;
    listMaxHeight?: number;
    volume?: number;
    mutex?: boolean;
    preload?: "auto" | "metadata" | "none";
  }>;
  export default APlayer;
}
