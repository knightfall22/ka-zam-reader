<script lang="ts">
  import { read } from "$app/server";
  import { readingState } from "$lib/state/state.svelte";
  import ArrowBigRightDash from "lucide-svelte/icons/arrow-big-right-dash";
  import ArrowBigLeftDash from "lucide-svelte/icons/arrow-big-left-dash";
  import Expand from "lucide-svelte/icons/expand";
  import Minimize from "lucide-svelte/icons/minimize";
  import X from "lucide-svelte/icons/x";
  import { goto } from "$app/navigation";

  function reload() {
    window.location.reload();
  }

  function next() {
    if (readingState.page + 1 < readingState.length) {
      readingState.page += 1;
    }
    document.body.scrollIntoView({ behavior: "smooth" });
  }

  function prev() {
    if (readingState.page > 0) {
      readingState.page -= 1;
    }
    document.body.scrollIntoView({ behavior: "smooth" });
  }

  function onKeyDown(e: KeyboardEvent) {
    switch (e.key) {
      case "ArrowLeft":
        e.preventDefault();
        prev();

        break;
      case "ArrowRight":
        e.preventDefault();
        next();
        break;

      case "Escape":
        e.preventDefault();
        history.back();
        break;

      case "m":
        e.preventDefault();
        isFullWidth = !isFullWidth;
        break;
    }
  }

  function cancel() {
    goto("/");
  }

  function changePage(e: Event) {
    const input = e.target as HTMLInputElement;

    if (isNaN(Number(input.value))) {
      input.value = readingState.page.toString();
    } else if (Number(input.value) > readingState.length) {
      readingState.page = readingState.length - 1;
      input.value = readingState.length.toString();
    } else if (Number(input.value) < 0) {
      readingState.page = 0;
      input.value = readingState.length.toString();
    } else if (input.value === "") {
      readingState.page = 0;
      input.value = readingState.page.toString();
    }
  }

  let isHovered = false;
  let isFullWidth = false;
</script>

<div class="flex flex-col items-center justify-center relative">
  {#if readingState.isReading}
    <img
      id="photo"
      src="/image/{readingState.page}"
      alt="No comic page found"
      class="transition-all duration-300 shadow-2xl rounded-md"
      class:w-auto={!isFullWidth}
      class:h-[100vh]={!isFullWidth}
      class:w-full={isFullWidth}
      class:h-auto={isFullWidth}
      onmouseover={() => (isHovered = true)}
      onmouseleave={() => (isHovered = false)}
      onfocus={() => (isHovered = true)}
      onblur={() => (isHovered = false)}
    />
  {/if}

  <!-- Toolbar -->
  <div
    class="absolute bottom-2 left-1/2 -translate-x-1/2 flex items-center space-x-2 bg-gray-800 text-white rounded-full p-2 transition-opacity duration-300"
    onmouseover={() => (isHovered = true)}
    onmouseleave={() => (isHovered = false)}
    onfocus={() => (isHovered = true)}
    onblur={() => (isHovered = false)}
    class:opacity-95={isHovered}
    class:opacity-0={!isHovered}
    role="tooltip"
  >
    <button
      class="p-2 hover:bg-gray-700 hover:rounded-tl-full hover:rounded-bl-full rounded cursor-pointer focus:bg-gray-700 focus:rounded-tl-full focus:rounded-bl-full"
      onclick={() => (isFullWidth = !isFullWidth)}
    >
      {#if isFullWidth}
        <Minimize size="18" />
      {:else}
        <Expand size="17.5" />
      {/if}
    </button>

    <button
      class="p-2 hover:bg-gray-700 rounded cursor-pointer"
      onclick={() => history.back()}
    >
      <X size="20" />
    </button>
    <button class="p-2 hover:bg-gray-700 rounded cursor-pointer" onclick={prev}>
      <ArrowBigLeftDash />
    </button>
    <div class="flex items-center space-x-2">
      <input
        type="number"
        class="w-12 px-1.5 bg-gray-900 text-center text-white appearance-none [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
        bind:value={readingState.page}
        oninput={changePage}
        min="0"
        max={readingState.length - 1}
      />
      <span class="px-2 py-1 rounded text-xs">
        ({readingState.page} of {readingState.length})
      </span>
    </div>
    <button
      class="p-2 hover:bg-gray-700 hover:rounded-br-full hover:rounded-tr-full rounded cursor-pointer focus:bg-gray-700 focus:rounded-br-full focus:rounded-tr-full"
      onclick={next}
    >
      <ArrowBigRightDash />
    </button>
  </div>
</div>

<svelte:window on:keydown={onKeyDown} />
