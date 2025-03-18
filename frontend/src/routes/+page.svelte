<script lang="ts">
  import { Length, LoadBook, SelectFile } from "$lib/wailsjs/go/main/App";
  import { readingState } from "$lib/state/state.svelte";
  import { goto } from "$app/navigation";
  import FolderOpen from "lucide-svelte/icons/folder-open";
  import BookOpen from "lucide-svelte/icons/book-open";
  import MessageCircle from "lucide-svelte/icons/message-circle";
  import Loader from "lucide-svelte/icons/loader";

  let loading: boolean = $state(false);

  function reload() {
    window.location.reload();
  }

  function selectFile() {
    SelectFile().then((filename) => {
      loading = true;
      readingState.isReading = false;
      LoadBook(filename).then(() => {
        loading = false;
        readingState.isReading = true;
        readingState.page = 0;
        Length().then((res) => {
          readingState.length = Number(res);
          goto("/reader");
        });
      });
    });
  }

  function toBook() {
    goto("/reader");
  }
</script>

<div
  class="flex flex-col items-center justify-center min-h-screen bg-[#f8f9fa] text-gray-900 text-center px-4"
>
  <!-- Title & Subtitle -->
  <h1 class="text-4xl font-bold mb-2 flex items-center gap-2">
    Ka-Zam Lite <MessageCircle class="inline-block self-start" />
  </h1>
  <p class="text-lg text-gray-600 mb-6">Lightweight CBR Reader</p>

  <!-- Buttons -->
  <div class="space-x-3 flex">
    <button
      class="px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded text-white font-medium transition flex items-center gap-2"
      onclick={selectFile}
    >
      <FolderOpen class="w-5 h-5" /> Open
    </button>

    {#if readingState.isReading}
      <button
        class="px-6 py-2 bg-green-600 hover:bg-green-700 rounded text-white font-medium transition flex items-center gap-2"
        onclick={toBook}
      >
        <BookOpen class="w-5 h-5" /> Open Previous
      </button>
    {/if}
  </div>

  {#if loading}
    <div class="mt-4 flex items-center gap-2 text-gray-500">
      <Loader class="w-6 h-6 animate-spin" /> Loading...
    </div>
  {/if}
</div>
