class PIrateRFController {
  constructor() {
    this.ws = null;
    this.isExecuting = false;
    this.wasConnected = false;
    this.currentAudioPath = null;
    this.currentPlayButton = null;

    // Initialize centralized state object
    this.state = {
      modulename: "pifmrds",

      pifmrds: {
        freq: "431",
        audio: "",
        pi: "",
        ps: "",
        rt: "",
        timeout: "0",
        playOnce: false,
        introOutroToggled: false,
        introSelect: "",
        outroSelect: "",
      },

      morse: {
        freq: "431000000",
        rate: "20",
        message: "HACK THE PLANET",
      },

      tune: {
        freq: "431000000",
        exitImmediate: false,
        ppm: "",
      },

      spectrumpaint: {
        frequency: "431000000",
        pictureFile: "",
        excursion: "50000",
      },

      pichirp: {
        frequency: "431000000",
        bandwidth: "1000000",
        time: "5.0",
      },
    };

    this.initializeElements();
    this.isDebugMode = this.checkDebugMode();
    this.bindEvents();
    this.onModuleChange(); // Initialize form visibility for default selected
    // module
    this.loadAudioFiles(false).then(() => {
      // Restore state after audio files are loaded
      this.restoreState();
    });
    this.connect();
  }

  // Build a proper file path using env.js config
  buildFilePath(fileName, type, forServer = false) {
    // Extract just the filename (handle cases where API returns paths)
    const justFilename = fileName.includes("/")
      ? fileName.split("/").pop()
      : fileName;

    // Build the full path using config
    if (type === "sfx") {
      if (forServer) {
        return `${window.PIrateRFConfig.serverPaths.audioSFX}/${justFilename}`;
      }
      return `${window.PIrateRFConfig.paths.files}/${window.PIrateRFConfig.directories.audioSFX}/${justFilename}`;
    } else if (type === "uploads") {
      if (forServer) {
        return `${window.PIrateRFConfig.serverPaths.audioUploads}/${justFilename}`;
      }
      return `${window.PIrateRFConfig.paths.files}/${window.PIrateRFConfig.directories.audioUploads}/${justFilename}`;
    } else if (type === "imageUploads") {
      if (forServer) {
        return `${window.PIrateRFConfig.serverPaths.imageUploads}/${justFilename}`;
      }
      return `${window.PIrateRFConfig.paths.files}/${window.PIrateRFConfig.directories.imageUploads}/${justFilename}`;
    }

    // If no type specified, assume it's already a full path
    // (like from dropdown)
    return fileName;
  }

  checkDebugMode() {
    const urlParams = new URLSearchParams(window.location.search);
    const debug = urlParams.has("debug");
    if (debug) {
      this.log("🐛 DEBUG MODE ENABLED");
    } else {
      this.log("👤 USER MODE");
    }
    return debug;
  }

  initializeElements() {
    this.titleEl = document.getElementById("title");
    this.statusBar = document.getElementById("statusBar");
    this.statusText = document.getElementById("statusText");
    this.moduleSelect = document.getElementById("moduleSelect");
    this.startBtn = document.getElementById("startBtn");
    this.stopBtn = document.getElementById("stopBtn");
    this.outputContent = document.getElementById("outputContent");
    this.controlPanel = document.getElementById("controlPanel");

    // Module forms
    this.pifmrdsForm = document.getElementById("pifmrdsForm");
    this.morseForm = document.getElementById("morseForm");
    this.tuneForm = document.getElementById("tuneForm");
    this.pichirpForm = document.getElementById("pichirpForm");

    // PIFMRDS form inputs
    this.freqInput = document.getElementById("freq");
    this.audioInput = document.getElementById("audio");
    this.playModeToggle = document.getElementById("playModeToggle");
    this.introOutroToggle = document.getElementById("introOutroToggle");
    this.introOutroControls = document.getElementById("introOutroControls");
    this.introSelect = document.getElementById("introSelect");
    this.outroSelect = document.getElementById("outroSelect");
    this.playIntroBtn = document.getElementById("playIntroBtn");
    this.playOutroBtn = document.getElementById("playOutroBtn");
    this.piInput = document.getElementById("pi");
    this.psInput = document.getElementById("ps");
    this.rtInput = document.getElementById("rt");
    this.ppmInput = document.getElementById("ppm");
    this.timeoutInput = document.getElementById("timeout");

    // MORSE form inputs
    this.morseFreqInput = document.getElementById("morseFreq");
    this.morseRateInput = document.getElementById("morseRate");
    this.morseMessageInput = document.getElementById("morseMessage");

    // TUNE form inputs
    this.tuneFreqInput = document.getElementById("tuneFreq");
    this.tuneExitImmediateInput = document.getElementById("tuneExitImmediate");
    this.tunePPMInput = document.getElementById("tunePPM");

    // SPECTRUMPAINT form inputs
    this.spectrumpaintForm = document.getElementById("spectrumpaintForm");
    this.spectrumpaintFreqInput = document.getElementById("spectrumpaintFreq");
    this.pictureFileInput = document.getElementById("pictureFile");
    this.excursionInput = document.getElementById("excursion");

    // PICHIRP form inputs
    this.pichirpFreqInput = document.getElementById("pichirpFreq");
    this.pichirpBandwidthInput = document.getElementById("pichirpBandwidth");
    this.pichirpTimeInput = document.getElementById("pichirpTime");
    this.refreshImageBtn = document.getElementById("refreshImageBtn");
    this.editImageBtn = document.getElementById("editImageBtn");
    this.imageSelectBtn = document.getElementById("imageSelectBtn");
    this.imageFile = document.getElementById("imageFile");
    this.imageUploadStatus = document.getElementById("imageUploadStatus");

    // Audio file dropdown
    this.refreshAudioBtn = document.getElementById("refreshAudioBtn");
    this.editAudioBtn = document.getElementById("editAudioBtn");

    // Modal elements
    this.fileEditModal = document.getElementById("fileEditModal");
    this.modalCloseBtn = document.getElementById("modalCloseBtn");
    this.modalCancelBtn = document.getElementById("modalCancelBtn");
    this.renameFileBtn = document.getElementById("renameFileBtn");
    this.editFileName = document.getElementById("editFileName");
    this.deleteFileBtn = document.getElementById("deleteFileBtn");

    // Image modal elements
    this.imageEditModal = document.getElementById("imageEditModal");
    this.imageModalCloseBtn = document.getElementById("imageModalCloseBtn");
    this.imageModalCancelBtn = document.getElementById("imageModalCancelBtn");
    this.renameImageBtn = document.getElementById("renameImageBtn");
    this.editImageName = document.getElementById("editImageName");
    this.deleteImageBtn = document.getElementById("deleteImageBtn");

    // Playlist modal elements
    this.playlistBtn = document.getElementById("playlistBtn");
    this.playlistModal = document.getElementById("playlistModal");
    this.playlistModalCloseBtn = document.getElementById(
      "playlistModalCloseBtn"
    );
    this.sfxFileList = document.getElementById("sfxFileList");
    this.uploadedFileList = document.getElementById("uploadedFileList");
    this.playlistItems = document.getElementById("playlistItems");
    this.clearPlaylistBtn = document.getElementById("clearPlaylistBtn");
    this.playlistName = document.getElementById("playlistName");
    this.createPlaylistBtn = document.getElementById("createPlaylistBtn");
    this.playlistError = document.getElementById("playlistError");

    // Audio playback elements
    this.audioPlayer = document.getElementById("audioPlayer");
    this.playAudioBtn = document.getElementById("playAudioBtn");

    // Current file being edited
    this.currentEditingFile = null;
    this.currentEditingImageFile = null;
    this.imageUploadStatusTimeout = null;

    // Playlist functionality
    this.playlist = [];

    // Audio file upload/record elements
    this.audioFileInput = document.getElementById("audioFile");
    this.fileSelectBtn = document.getElementById("fileSelectBtn");
    this.uploadStatus = document.getElementById("uploadStatus");
    this.recordBtn = document.getElementById("recordBtn");
    this.recordStatus = document.getElementById("recordStatus");

    // Recording state
    this.mediaRecorder = null;
    this.audioChunks = [];
    this.isRecording = false;

    // Error notification elements
    this.errorNotification = document.getElementById("errorNotification");
    this.errorMessage = document.getElementById("errorMessage");
    this.errorClose = document.getElementById("errorClose");
    this.errorTimeout = null;

    // Status message timeouts
    this.uploadStatusTimeout = null;
    this.recordStatusTimeout = null;

    // Loading screen elements
    this.loadingOverlay = document.getElementById("loadingOverlay");
    this.loadingText = document.querySelector(".loading-text");
  }

  bindEvents() {
    this.moduleSelect.addEventListener("change", () => {
      this.onModuleChange();
      this.saveState();
    });
    this.startBtn.addEventListener("click", () => this.startExecution());
    this.stopBtn.addEventListener("click", () => this.stopExecution());

    // Play mode toggle
    this.playModeToggle.addEventListener("click", () => {
      this.togglePlayMode();
      this.saveState();
    });

    // Intro/Outro toggle
    this.introOutroToggle.addEventListener("click", () => {
      this.toggleIntroOutro();
      this.saveState();
    });

    // Intro/Outro play buttons and select change handlers
    this.introSelect.addEventListener("change", () => {
      this.onIntroOutroChange();
      this.saveState();
    });
    this.outroSelect.addEventListener("change", () => {
      this.onIntroOutroChange();
      this.saveState();
    });
    this.playIntroBtn.addEventListener("click", () =>
      this.playIntroOutro("intro")
    );
    this.playOutroBtn.addEventListener("click", () =>
      this.playIntroOutro("outro")
    );

    // Random generation buttons
    document
      .getElementById("randomPi")
      .addEventListener("click", () => this.generateRandomPI());
    document
      .getElementById("randomPs")
      .addEventListener("click", () => this.generateRandomPS());
    document
      .getElementById("randomRt")
      .addEventListener("click", () => this.generateRandomRT());

    // Audio file dropdown events
    this.refreshAudioBtn.addEventListener("click", () => this.loadAudioFiles());
    this.editAudioBtn.addEventListener("click", () => this.openEditModal());
    this.audioInput.addEventListener("change", () => {
      this.onAudioFileChange();
      this.saveState();
    });
    this.playAudioBtn.addEventListener("click", () => this.playSelectedAudio());

    // Modal events
    this.modalCloseBtn.addEventListener("click", () => this.closeEditModal());
    this.modalCancelBtn.addEventListener("click", () => this.closeEditModal());
    this.renameFileBtn.addEventListener("click", () => this.renameFile());
    this.deleteFileBtn.addEventListener("click", () => this.deleteFile());

    // Image modal events
    this.imageModalCloseBtn.addEventListener("click", () =>
      this.closeImageEditModal()
    );
    this.imageModalCancelBtn.addEventListener("click", () =>
      this.closeImageEditModal()
    );
    this.renameImageBtn.addEventListener("click", () => this.renameImageFile());
    this.deleteImageBtn.addEventListener("click", () => this.deleteImageFile());

    // Image upload handling
    this.imageFile.addEventListener("change", () => this.uploadImage());

    // Playlist modal events
    this.playlistBtn.addEventListener("click", () => this.openPlaylistModal());
    this.playlistModalCloseBtn.addEventListener("click", () =>
      this.closePlaylistModal()
    );
    this.clearPlaylistBtn.addEventListener("click", () => this.clearPlaylist());
    this.createPlaylistBtn.addEventListener("click", () =>
      this.createPlaylist()
    );
    this.playlistName.addEventListener("input", () =>
      this.validatePlaylistCreation()
    );

    // Close modal when clicking outside
    this.fileEditModal.addEventListener("click", (e) => {
      if (e.target === this.fileEditModal) {
        this.closeEditModal();
      }
    });

    this.playlistModal.addEventListener("click", (e) => {
      if (e.target === this.playlistModal) {
        this.closePlaylistModal();
      }
    });

    // File upload events
    this.fileSelectBtn.addEventListener("click", () =>
      this.audioFileInput.click()
    );
    this.audioFileInput.addEventListener("change", () => this.onFileSelected());

    // Recording events
    this.recordBtn.addEventListener("click", () => this.toggleRecording());

    // Error notification events
    this.errorClose.addEventListener("click", () =>
      this.hideErrorNotification()
    );

    // Validate required fields for all modules
    [
      // PIFMRDS inputs
      this.freqInput,
      this.audioInput,
      // MORSE inputs
      this.morseFreqInput,
      this.morseRateInput,
      this.morseMessageInput,
      // TUNE inputs
      this.tuneFreqInput,
      this.tuneExitImmediateInput,
      this.tunePPMInput,
    ].forEach((input) => {
      if (input) {
        input.addEventListener("input", () => this.validateForm());
        input.addEventListener("change", () => this.validateForm());
      }
    });

    // Window resize handler for execution mode
    window.addEventListener("resize", () => {
      if (this.isExecuting) {
        setTimeout(() => this.adjustSystemOutputHeight(), 100);
      }
    });

    // Form input events for state saving
    this.freqInput.addEventListener("input", () => this.saveState());
    this.piInput.addEventListener("input", () => this.saveState());
    this.psInput.addEventListener("input", () => this.saveState());
    this.rtInput.addEventListener("input", () => this.saveState());
    document
      .getElementById("timeout")
      .addEventListener("input", () => this.saveState());

    // MORSE module form events
    document
      .getElementById("morseFreq")
      .addEventListener("input", () => this.saveState());
    document
      .getElementById("morseRate")
      .addEventListener("input", () => this.saveState());
    document
      .getElementById("morseMessage")
      .addEventListener("input", () => this.saveState());

    // TUNE module form events
    document
      .getElementById("tuneFreq")
      .addEventListener("input", () => this.saveState());
    document
      .getElementById("tuneExitImmediate")
      .addEventListener("change", () => this.saveState());
    document
      .getElementById("tunePPM")
      .addEventListener("input", () => this.saveState());

    // SPECTRUMPAINT module form events
    this.spectrumpaintFreqInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.pictureFileInput.addEventListener("change", () => {
      this.onImageFileChange();
      this.saveState();
      this.validateForm();
    });
    this.excursionInput.addEventListener("input", () => this.saveState());

    // SPECTRUMPAINT image control buttons
    this.refreshImageBtn.addEventListener("click", () => this.loadImageFiles());
    this.editImageBtn.addEventListener("click", () =>
      this.openImageEditModal()
    );
    this.imageSelectBtn.addEventListener("click", () => this.imageFile.click());

    // PICHIRP module form events
    this.pichirpFreqInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.pichirpBandwidthInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.pichirpTimeInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
  }

  connect() {
    // Use the current page's host and port for WebSocket connection
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.hostname;
    const port =
      window.location.port ||
      (window.location.protocol === "https:" ? "443" : "80");
    const wsUrl = `${protocol}//${host}:${port}/ws`;

    this.connectToUrl(wsUrl);
  }

  connectToUrl(wsUrl) {
    try {
      this.statusText.textContent = `🔌 Connecting...`;
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        this.wasConnected = true;
        this.titleEl.className = "title connected";
        this.statusText.textContent = "Idle";
        this.clearWebSocketErrors(); // Clear any WebSocket error notifications
      };

      this.ws.onclose = (event) => {
        this.titleEl.className = "title disconnected";

        // Only show disconnected status if we were actually connected
        if (this.wasConnected) {
          this.statusText.textContent = "🔌 Disconnected";
          this.showErrorNotification(
            "🔌 WebSocket disconnected - reconnecting...",
            "websocket-disconnected"
          );
          this.wasConnected = false;
        } else {
          this.statusText.textContent = "❌ Connection failed";
          this.showErrorNotification(
            "❌ WebSocket connection failed - retrying...",
            "websocket"
          );
        }

        // Try to reconnect after 3 seconds
        setTimeout(() => this.connect(), 3000);
      };

      this.ws.onerror = (error) => {
        if (!this.wasConnected) {
          this.statusText.textContent = "❌ Connection failed";
          this.showErrorNotification(
            "❌ WebSocket error occurred",
            "websocket"
          );
        }
      };

      this.ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        this.handleMessage(data);
      };
    } catch (error) {
      this.statusText.textContent = "❌ Connection failed";
      this.showErrorNotification(
        `❌ Connection error: ${error.message}`,
        "websocket"
      );
      setTimeout(() => this.connect(), 3000);
    }
  }

  handleMessage(message) {
    if (this.isDebugMode) {
      this.log("📨 RECEIVED: " + JSON.stringify(message, null, 2), "receive");
    }

    switch (message.type) {
      case "rpitx.execution.started":
        this.onExecutionStarted(message.data);
        break;
      case "rpitx.execution.stopped":
        this.onExecutionStopped(message.data);
        break;
      case "rpitx.execution.error":
        this.onExecutionError(message.data);
        break;
      case "rpitx.execution.output-line":
        this.onOutputLine(message.data);
        break;
      case "file.rename.success":
        // Check if it's an image or audio file based on the file path
        if (
          message.data.fileName &&
          message.data.fileName.includes("/images/")
        ) {
          this.onImageFileRenameSuccess(message.data);
        } else {
          this.onFileRenameSuccess(message.data);
        }
        break;
      case "file.rename.error":
        // Check if it's an image or audio file based on the file path
        if (
          message.data.fileName &&
          message.data.fileName.includes("/images/")
        ) {
          this.onImageFileRenameError(message.data);
        } else {
          this.onFileRenameError(message.data);
        }
        break;
      case "file.delete.success":
        // Check if it's an image or audio file based on the file path
        if (
          message.data.fileName &&
          message.data.fileName.includes("/images/")
        ) {
          this.onImageFileDeleteSuccess(message.data);
        } else {
          this.onFileDeleteSuccess(message.data);
        }
        break;
      case "file.delete.error":
        // Check if it's an image or audio file based on the file path
        if (
          message.data.fileName &&
          message.data.fileName.includes("/images/")
        ) {
          this.onImageFileDeleteError(message.data);
        } else {
          this.onFileDeleteError(message.data);
        }
        break;
      case "audio.playlist.create.success":
        this.onPlaylistCreateSuccess(message.data);
        break;
      case "audio.playlist.create.error":
        this.onPlaylistCreateError(message.data);
        break;
    }
  }

  onModuleChange() {
    const module = this.moduleSelect.value;

    // Hide all module forms
    this.pifmrdsForm.classList.add("hidden");
    this.morseForm.classList.add("hidden");
    this.tuneForm.classList.add("hidden");
    this.spectrumpaintForm.classList.add("hidden");
    this.pichirpForm.classList.add("hidden");

    // Show the selected module form
    switch (module) {
      case "pifmrds":
        this.pifmrdsForm.classList.remove("hidden");
        this.initializeRandomValues();
        break;
      case "morse":
        this.morseForm.classList.remove("hidden");
        break;
      case "tune":
        this.tuneForm.classList.remove("hidden");
        break;
      case "spectrumpaint":
        this.spectrumpaintForm.classList.remove("hidden");
        this.loadImageFiles(false);
        break;
      case "pichirp":
        this.pichirpForm.classList.remove("hidden");
        break;
    }

    this.validateForm();
  }

  initializeRandomValues() {
    // Only initialize if fields are empty
    if (!this.piInput.value) {
      this.generateRandomPI();
    }
    if (!this.psInput.value) {
      this.generateRandomPS();
    }
    if (!this.rtInput.value) {
      this.generateRandomRT();
    }
  }

  generateRandomPI() {
    // Generate 4 random hex digits - exactly like backend GetRandomHex(4)
    const length = 4;
    let maxVal = 1;
    for (let i = 0; i < length; i++) {
      maxVal *= 16;
    }
    let hexStr = Math.floor(Math.random() * maxVal)
      .toString(16)
      .toUpperCase();
    // Pad with leading zeros to ensure exact length
    while (hexStr.length < length) {
      hexStr = "0" + hexStr;
    }
    this.piInput.value = hexStr;
  }

  generateRandomPS() {
    // Generate random alphanumeric string - exactly like backend GetRandomAlphanumeric(8)
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    const length = 8;
    let result = "";
    for (let i = 0; i < length; i++) {
      result += chars[Math.floor(Math.random() * chars.length)];
    }
    this.psInput.value = result;
  }

  generateRandomRT() {
    // Generate random string in range - exactly like backend GetRandomStringInRange(10, 64)
    const chars =
      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789 ";
    const minLength = 10;
    const maxLength = 64;
    const length =
      Math.floor(Math.random() * (maxLength - minLength + 1)) + minLength;
    let result = "";
    for (let i = 0; i < length; i++) {
      result += chars[Math.floor(Math.random() * chars.length)];
    }
    this.rtInput.value = result;
  }

  setExecutionMode(isExecuting) {
    this.isExecuting = isExecuting;

    if (isExecuting) {
      // Set executing mode
      document.body.classList.add("executing");
      this.titleEl.textContent = "⚡️ PIrateRF ⚡️";
      this.statusBar.classList.add("executing");
      this.startBtn.classList.add("hidden");
      this.stopBtn.classList.remove("hidden");

      // Scroll to top and calculate system output height
      window.scrollTo(0, 0);
      setTimeout(() => this.adjustSystemOutputHeight(), 100);
    } else {
      // Set idle mode
      document.body.classList.remove("executing");
      this.titleEl.textContent = "🏴‍☠️ PIrateRF 🏴‍☠️";
      this.statusBar.classList.remove("executing");
      this.statusText.textContent = "Idle";
      this.stopBtn.classList.add("hidden");
      this.startBtn.classList.remove("hidden");

      // Reset system output height
      document.getElementById("systemOutput").style.height = "";
    }

    this.validateForm();
  }

  adjustSystemOutputHeight() {
    const systemOutput = document.getElementById("systemOutput");
    const container = document.querySelector(".container");

    // Get viewport height
    const viewportHeight = window.innerHeight;

    // Calculate used space by other elements
    const header = document.querySelector(".header");
    const statusBar = document.querySelector(".status-bar");
    const controlPanel = document.querySelector(".control-panel");

    const headerHeight = header ? header.offsetHeight : 0;
    const statusHeight = statusBar ? statusBar.offsetHeight : 0;
    const controlHeight = controlPanel ? controlPanel.offsetHeight : 0;

    // Get container padding
    const containerStyle = window.getComputedStyle(container);
    const containerPadding =
      parseFloat(containerStyle.paddingTop) +
      parseFloat(containerStyle.paddingBottom);

    // Minimal buffer to maximize system log space
    const margins = 20; // minimal margin between elements
    const extraBuffer = 10; // small buffer to ensure header stays visible

    // Calculate available height for system output
    const usedSpace =
      headerHeight +
      statusHeight +
      controlHeight +
      containerPadding +
      margins +
      extraBuffer;
    const availableHeight = viewportHeight - usedSpace;

    // Set the height
    const finalHeight = Math.max(availableHeight, 150);
    systemOutput.style.height = `${finalHeight}px`;

    // Only log viewport calculations in debug mode
    if (this.isDebugMode) {
      this.log(
        `📐 VH:${viewportHeight} H:${headerHeight} S:${statusHeight} C:${controlHeight} P:${containerPadding} → ${finalHeight}px`,
        "system"
      );
    }
  }

  validateForm() {
    const module = this.moduleSelect.value;
    let isValid = false;

    switch (module) {
      case "pifmrds":
        isValid = module && this.freqInput.value && this.audioInput.value;
        break;
      case "morse":
        isValid =
          module &&
          this.morseFreqInput.value &&
          this.morseRateInput.value &&
          this.morseMessageInput.value.trim();
        break;
      case "tune":
        isValid = module && this.tuneFreqInput.value;
        break;
      case "spectrumpaint":
        isValid =
          module &&
          this.spectrumpaintFreqInput.value &&
          this.pictureFileInput.value;
        break;
      case "pichirp":
        isValid =
          module &&
          this.pichirpFreqInput.value &&
          this.pichirpBandwidthInput.value &&
          this.pichirpTimeInput.value;
        break;
      default:
        isValid = false;
    }

    this.startBtn.disabled = !isValid || this.isExecuting;
  }

  startExecution() {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log("❌ WebSocket not connected", "system");
      return;
    }

    const module = this.moduleSelect.value;
    let args = {};
    let timeout = 0;

    switch (module) {
      case "pifmrds":
        args = {
          freq: parseFloat(this.freqInput.value),
          audio: this.audioInput.value,
        };
        // Add optional fields only if they have values
        if (this.piInput.value.trim()) {
          args.pi = this.piInput.value.trim();
        }
        if (this.psInput.value.trim()) {
          args.ps = this.psInput.value.trim();
        }
        if (this.rtInput.value.trim()) {
          args.rt = this.rtInput.value.trim();
        }
        if (this.ppmInput.value.trim()) {
          args.ppm = parseFloat(this.ppmInput.value);
        }
        timeout =
          this.timeoutInput.value === ""
            ? 30
            : parseInt(this.timeoutInput.value);
        break;

      case "morse":
        args = {
          frequency: parseFloat(this.morseFreqInput.value),
          rate: parseInt(this.morseRateInput.value),
          message: this.morseMessageInput.value.trim(),
        };
        timeout = 0; // No timeout for morse by default
        break;

      case "tune":
        args = {
          frequency: parseFloat(this.tuneFreqInput.value),
        };
        if (this.tuneExitImmediateInput.checked) {
          args.exitImmediate = true;
        }
        if (this.tunePPMInput.value.trim()) {
          args.ppm = parseFloat(this.tunePPMInput.value);
        }
        timeout = 0; // No timeout for tune by default
        break;

      case "spectrumpaint":
        args = {
          frequency: parseFloat(this.spectrumpaintFreqInput.value),
          pictureFile: this.pictureFileInput.value,
        };
        if (this.excursionInput.value.trim()) {
          args.excursion = parseFloat(this.excursionInput.value);
        }
        timeout = 0; // No timeout for spectrumpaint by default
        break;

      case "pichirp":
        args = {
          frequency: parseFloat(this.pichirpFreqInput.value),
          bandwidth: parseFloat(this.pichirpBandwidthInput.value),
          time: parseFloat(this.pichirpTimeInput.value),
        };
        timeout = 0; // No timeout for pichirp by default
        break;
    }

    const message = {
      type: "rpitx.execution.start",
      data: {
        moduleName: module,
        args: args,
        timeout: timeout,
        playOnce:
          module === "pifmrds"
            ? this.playModeToggle.classList.contains("play-once")
            : false,
        intro:
          module === "pifmrds" &&
          this.introOutroToggle.classList.contains("active")
            ? this.introSelect.value || null
            : null,
        outro:
          module === "pifmrds" &&
          this.introOutroToggle.classList.contains("active")
            ? this.outroSelect.value || null
            : null,
      },
      id: this.generateUUID(),
    };

    if (this.isDebugMode) {
      this.log("📤 SENDING: " + JSON.stringify(message, null, 2), "send");
    }
    this.ws.send(JSON.stringify(message));
  }

  stopExecution() {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log("❌ WebSocket not connected", "system");
      return;
    }

    const message = {
      type: "rpitx.execution.stop",
      data: {},
      id: this.generateUUID(),
    };

    if (this.isDebugMode) {
      this.log("📤 SENDING: " + JSON.stringify(message, null, 2), "send");
    }
    this.ws.send(JSON.stringify(message));
  }

  onExecutionStarted(data) {
    this.setExecutionMode(true);

    // Format as command line dynamically
    let cmdLine = data.moduleName;
    const args = data.args;

    for (const [key, value] of Object.entries(args)) {
      if (value !== null && value !== undefined && value !== "") {
        // Quote values that might contain spaces
        const needsQuotes =
          typeof value === "string" &&
          (value.includes(" ") || value.length > 20);
        const formattedValue = needsQuotes ? `"${value}"` : value;
        cmdLine += ` -${key} ${formattedValue}`;
      }
    }

    this.statusText.textContent = `Executing: ${cmdLine}`;

    // Show single consolidated execution message
    let executionMsg = `🚀 EXECUTION STARTED: ${data.moduleName.toUpperCase()} ${JSON.stringify(
      data.args
    )} (triggered by ${data.initiatingClientId})`;

    this.log(executionMsg, "system");

    if (this.isDebugMode) {
      this.log(`⚙️ Args: ${JSON.stringify(data.args)}`, "system");
    }
  }

  onExecutionStopped(data) {
    this.setExecutionMode(false);

    this.log("🛑 EXECUTION STOPPED", "system");
    if (this.isDebugMode) {
      this.log(`Client: ${data.stoppingClientId}`, "system");
    }
  }

  onExecutionError(data) {
    this.setExecutionMode(false);

    this.log(`❌ EXECUTION ERROR: ${data.error}`, "system");
    this.log(`Message: ${data.message}`, "system");
  }

  onOutputLine(data) {
    const prefix = data.type.toUpperCase();
    this.log(`[${prefix}] ${data.line}`, "output");
  }

  log(message, type = "system") {
    const entry = document.createElement("div");
    entry.className = `log-entry log-${type}`;
    entry.textContent = `[${new Date().toLocaleTimeString()}] ${message}`;

    this.outputContent.appendChild(entry);

    // Force autoscroll to bottom with animation frame timing
    requestAnimationFrame(() => {
      this.outputContent.scrollTop = this.outputContent.scrollHeight;
    });

    // Keep only last 1000 entries
    if (this.outputContent.children.length > 1000) {
      this.outputContent.removeChild(this.outputContent.firstChild);
    }
  }

  onFileSelected() {
    const hasFile = this.audioFileInput.files.length > 0;

    if (hasFile) {
      const file = this.audioFileInput.files[0];
      this.uploadStatus.textContent = `Selected: ${file.name}`;
      this.uploadStatus.className = "upload-status";

      // Auto-upload the file immediately
      this.uploadFile();
    } else {
      this.clearUploadStatus();
    }
  }

  async uploadFile() {
    const file = this.audioFileInput.files[0];
    if (!file) {
      this.setUploadStatus("No file selected", "error");
      return;
    }

    const formData = new FormData();
    formData.append("file", file);

    this.setUploadStatus("Uploading...", "uploading");

    try {
      const response = await this.customFetch(
        "/upload",
        {
          method: "POST",
          body: formData,
        },
        "Uploading file..."
      );

      if (!response.ok) {
        throw new Error(
          `Upload failed: ${response.status} ${response.statusText}`
        );
      }

      const result = await response.json();

      if (result.status === "success") {
        this.setUploadStatus(
          `Uploaded: ${result.original_filename}`,
          "success"
        );

        // Clear the file input after successful upload
        this.audioFileInput.value = "";

        // Refresh dropdown to show new file and auto-select it
        await this.loadAudioFiles();

        this.validateForm();
      } else {
        throw new Error("Upload failed: Invalid response");
      }
    } catch (error) {
      this.setUploadStatus(`Error: ${error.message}`, "error");
      this.log(`❌ Upload error: ${error.message}`, "system");
    } finally {
      // Upload is now automatic, no button to re-enable
    }
  }

  setUploadStatus(message, type) {
    this.uploadStatus.textContent = message;
    this.uploadStatus.className = `upload-status ${type}`;

    // Clear any existing timeout
    if (this.uploadStatusTimeout) {
      clearTimeout(this.uploadStatusTimeout);
    }

    // Auto-clear after 10 seconds for error and success messages
    if (type === "error" || type === "success") {
      this.uploadStatusTimeout = setTimeout(() => {
        this.clearUploadStatus();
      }, 10000);
    }
  }

  clearUploadStatus() {
    this.uploadStatus.textContent = "";
    this.uploadStatus.className = "upload-status";

    // Clear timeout if it exists
    if (this.uploadStatusTimeout) {
      clearTimeout(this.uploadStatusTimeout);
      this.uploadStatusTimeout = null;
    }
  }

  async uploadImage() {
    const file = this.imageFile.files[0];
    if (!file) {
      this.setImageUploadStatus("No file selected", "error");
      return;
    }

    const formData = new FormData();
    formData.append("file", file);

    this.setImageUploadStatus("Uploading...", "uploading");

    try {
      const response = await this.customFetch(
        "/upload",
        {
          method: "POST",
          body: formData,
        },
        "Uploading image..."
      );

      if (!response.ok) {
        throw new Error(
          `Upload failed: ${response.status} ${response.statusText}`
        );
      }

      const result = await response.json();

      if (result.status === "success") {
        this.setImageUploadStatus(
          `Uploaded: ${result.original_filename}`,
          "success"
        );

        // Clear the file input after successful upload
        this.imageFile.value = "";

        // Refresh dropdown to show new file and auto-select it
        await this.loadImageFiles();

        this.validateForm();
      } else {
        throw new Error("Upload failed: Invalid response");
      }
    } catch (error) {
      this.setImageUploadStatus(`Error: ${error.message}`, "error");
      this.log(`❌ Image upload error: ${error.message}`, "system");
    }
  }

  setImageUploadStatus(message, type) {
    this.imageUploadStatus.textContent = message;
    this.imageUploadStatus.className = `upload-status ${type}`;

    // Clear any existing timeout
    if (this.imageUploadStatusTimeout) {
      clearTimeout(this.imageUploadStatusTimeout);
    }

    // Auto-clear after 10 seconds for error and success messages
    if (type === "error" || type === "success") {
      this.imageUploadStatusTimeout = setTimeout(() => {
        this.clearImageUploadStatus();
      }, 10000);
    }
  }

  clearImageUploadStatus() {
    this.imageUploadStatus.textContent = "";
    this.imageUploadStatus.className = "upload-status";

    // Clear timeout if it exists
    if (this.imageUploadStatusTimeout) {
      clearTimeout(this.imageUploadStatusTimeout);
      this.imageUploadStatusTimeout = null;
    }
  }

  clearRecordStatus() {
    this.recordStatus.textContent = "";
    this.recordStatus.className = "record-status";

    // Clear timeout if it exists
    if (this.recordStatusTimeout) {
      clearTimeout(this.recordStatusTimeout);
      this.recordStatusTimeout = null;
    }
  }

  setRecordStatus(message, type) {
    this.recordStatus.textContent = message;
    this.recordStatus.className = `record-status ${type}`;

    // Clear any existing timeout
    if (this.recordStatusTimeout) {
      clearTimeout(this.recordStatusTimeout);
    }

    // Auto-clear after 10 seconds for error and success messages
    if (type === "error" || type === "success") {
      this.recordStatusTimeout = setTimeout(() => {
        this.clearRecordStatus();
      }, 10000);
    }
  }

  async toggleRecording() {
    if (this.isRecording) {
      this.stopRecording();
    } else {
      await this.startRecording();
    }
  }

  async startRecording() {
    try {
      // Check if MediaRecorder is supported
      if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
        throw new Error("MediaRecorder not supported in this browser");
      }

      if (this.isDebugMode) {
        this.log(`🎤 Requesting microphone access...`, "system");
      }
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

      if (this.isDebugMode) {
        this.log(
          `🎤 Microphone access granted, starting recording...`,
          "system"
        );
      }

      // Try formats in order of sox compatibility and browser support
      let options = {};
      let selectedFormat = null;

      const audioFormats = [
        { mimeType: "audio/webm;codecs=opus", extension: ".webm" },
        { mimeType: "audio/webm", extension: ".webm" },
        { mimeType: "audio/ogg;codecs=opus", extension: ".ogg" },
        { mimeType: "audio/ogg", extension: ".ogg" },
        { mimeType: "audio/mp4", extension: ".m4a" },
        { mimeType: "audio/wav", extension: ".wav" },
      ];

      // Find first supported format
      for (const format of audioFormats) {
        if (MediaRecorder.isTypeSupported(format.mimeType)) {
          selectedFormat = format;
          break;
        }
      }

      if (selectedFormat) {
        options.mimeType = selectedFormat.mimeType;
        this.recordingMimeType = selectedFormat.mimeType;
        this.recordingExtension = selectedFormat.extension;
        if (this.isDebugMode) {
          this.log(
            `🎤 Using format: ${selectedFormat.mimeType} → ${selectedFormat.extension}`,
            "system"
          );
        }
      } else {
        // Fallback to browser default (usually ogg for audio)
        this.recordingMimeType = "audio/ogg";
        this.recordingExtension = ".ogg";
        if (this.isDebugMode) {
          this.log(`🎤 Using browser default: audio/ogg → .ogg`, "system");
        }
      }

      this.mediaRecorder = new MediaRecorder(stream, options);
      this.audioChunks = [];

      this.mediaRecorder.ondataavailable = (event) => {
        this.audioChunks.push(event.data);
      };

      this.mediaRecorder.onstop = () => {
        this.processRecording();
        stream.getTracks().forEach((track) => track.stop()); // Stop microphone access
      };

      this.mediaRecorder.start();
      this.isRecording = true;

      // Update UI
      this.recordBtn.textContent = "🛑";
      this.recordBtn.classList.add("recording");
      this.setRecordStatus("Recording...", "recording");
    } catch (error) {
      this.log(
        `❌ Recording error details: ${error.name} - ${error.message}`,
        "system"
      );

      if (error.name === "NotFoundError") {
        this.setRecordStatus(`No microphone found`, "error");
      } else if (error.name === "NotAllowedError") {
        this.setRecordStatus(`Microphone permission denied`, "error");
      } else {
        this.setRecordStatus(`Error: ${error.message}`, "error");
      }
    }
  }

  stopRecording() {
    if (this.mediaRecorder && this.isRecording) {
      this.mediaRecorder.stop();
      this.isRecording = false;

      // Update UI
      this.recordBtn.textContent = "🎤";
      this.recordBtn.classList.remove("recording");
      this.setRecordStatus("Processing recording...", "uploading");
    }
  }

  async processRecording() {
    try {
      const audioBlob = new Blob(this.audioChunks, {
        type: this.recordingMimeType,
      });

      // Generate filename with unix timestamp and correct extension
      const unixTime = Math.floor(Date.now() / 1000);
      const filename = `recording_${unixTime}${this.recordingExtension}`;

      // Create form data and upload
      const formData = new FormData();
      formData.append("file", audioBlob, filename);

      this.setRecordStatus("Uploading recording...", "uploading");

      const response = await this.customFetch(
        "/upload",
        {
          method: "POST",
          body: formData,
        },
        "Uploading recording..."
      );

      if (!response.ok) {
        throw new Error(
          `Upload failed: ${response.status} ${response.statusText}`
        );
      }

      const result = await response.json();

      if (this.isDebugMode) {
        this.log(`📤 Upload response: ${JSON.stringify(result)}`, "system");
      }

      if (result.status === "success") {
        this.setRecordStatus(`Recorded and uploaded: ${filename}`, "success");

        // Refresh dropdown to show new file and auto-select it
        await this.loadAudioFiles();

        this.validateForm();
      } else {
        throw new Error("Upload failed: Invalid response");
      }
    } catch (error) {
      this.setRecordStatus(`Error: ${error.message}`, "error");
      this.log(`❌ Recording upload error: ${error.message}`, "system");
    }
  }

  async loadAudioFiles(selectLatest = true) {
    try {
      const response = await this.customFetch(
        window.PIrateRFConfig.paths.audioUploadFiles,
        {},
        "Loading audio files..."
      );
      if (!response.ok) {
        throw new Error(`Failed to load audio files: ${response.status}`);
      }

      const files = await response.json();

      if (this.isDebugMode) {
        this.log(`🔄 Loaded ${files.length} audio files`, "system");
      }

      // Sort by modTime (newest first)
      files.sort((a, b) => new Date(b.modTime) - new Date(a.modTime));

      // Clear existing options
      this.audioInput.innerHTML = "";

      // Filter audio files
      const audioFiles = files.filter(
        (file) => !file.isDir && file.name.endsWith(".wav")
      );

      if (audioFiles.length === 0) {
        // No audio files found
        const option = document.createElement("option");
        option.value = "";
        option.textContent = "No audio files";
        option.disabled = true;
        this.audioInput.appendChild(option);
      } else {
        // Add file options
        audioFiles.forEach((file) => {
          const option = document.createElement("option");
          option.value = this.buildFilePath(file.name, "uploads", true); // Server path for backend
          option.textContent = file.name;
          this.audioInput.appendChild(option);
        });

        // Try to restore saved audio file selection, otherwise select first (newest) file
        if (selectLatest) {
          // Select first (newest) file when selectLatest is true
          this.audioInput.selectedIndex = 0;
        } else {
          // Try to restore saved selection only when selectLatest is false (page load)
          this.selectSavedOrFirstAudioFile();
        }
        this.validateForm();
      }

      // Enable/disable edit button based on selection
      this.onAudioFileChange();
    } catch (error) {
      this.log(`❌ Failed to load audio files: ${error.message}`, "system");
    }
  }

  async loadImageFiles(selectLatest = true) {
    try {
      const response = await this.customFetch(
        window.PIrateRFConfig.paths.imageUploadFiles,
        {},
        "Loading image files..."
      );
      if (!response.ok) {
        throw new Error(`Failed to load image files: ${response.status}`);
      }

      const files = await response.json();

      if (this.isDebugMode) {
        this.log(`🔄 Loaded ${files.length} image files`, "system");
      }

      // Sort by modTime (newest first)
      files.sort((a, b) => new Date(b.modTime) - new Date(a.modTime));

      // Clear existing options
      this.pictureFileInput.innerHTML = "";

      // Filter image files (.Y files)
      const imageFiles = files.filter(
        (file) => !file.isDir && file.name.endsWith(".Y")
      );

      if (imageFiles.length === 0) {
        // No image files found
        const option = document.createElement("option");
        option.value = "";
        option.textContent = "No image files";
        option.disabled = true;
        this.pictureFileInput.appendChild(option);
      } else {
        // Add file options
        imageFiles.forEach((file) => {
          const option = document.createElement("option");
          option.value = this.buildFilePath(file.name, "imageUploads", true); // Server path for backend
          option.textContent = file.name;
          this.pictureFileInput.appendChild(option);
        });

        // Try to restore saved image file selection, otherwise select first (newest) file
        if (selectLatest) {
          this.pictureFileInput.selectedIndex = 0;
        } else {
          this.selectSavedOrFirstImageFile();
        }
        this.validateForm();
      }

      // Enable/disable edit button based on selection
      this.onImageFileChange();
    } catch (error) {
      this.log(`❌ Failed to load image files: ${error.message}`, "system");
    }
  }

  onImageFileChange() {
    const hasSelection =
      this.pictureFileInput.value && this.pictureFileInput.value !== "";
    this.editImageBtn.disabled = !hasSelection;
  }

  selectSavedOrFirstImageFile() {
    if (this.state.spectrumpaint && this.state.spectrumpaint.pictureFile) {
      const savedValue = this.state.spectrumpaint.pictureFile;
      for (let i = 0; i < this.pictureFileInput.options.length; i++) {
        if (this.pictureFileInput.options[i].value === savedValue) {
          this.pictureFileInput.selectedIndex = i;
          return;
        }
      }
    }
    if (
      this.pictureFileInput.options.length > 0 &&
      !this.pictureFileInput.options[0].disabled
    ) {
      this.pictureFileInput.selectedIndex = 0;
    }
  }

  onAudioFileChange() {
    const hasSelection = this.audioInput.value && this.audioInput.value !== "";
    this.editAudioBtn.disabled = !hasSelection;
    this.playAudioBtn.disabled = !hasSelection;
  }

  async openEditModal() {
    const selectedValue = this.audioInput.value;
    if (!selectedValue || selectedValue === "") {
      return;
    }

    // Extract filename from server path (selectedValue is a server path like "./files/audio/uploads/file.wav")
    const serverAudioPath = window.PIrateRFConfig.serverPaths.audioUploads;
    if (!selectedValue.startsWith(serverAudioPath)) {
      return;
    }

    this.currentEditingFile = selectedValue.replace(`${serverAudioPath}/`, "");

    // Set the filename in the input
    this.editFileName.value = this.currentEditingFile;

    // Show the modal
    this.fileEditModal.style.display = "flex";
  }

  closeEditModal() {
    this.fileEditModal.style.display = "none";
    this.currentEditingFile = null;
    this.editFileName.value = "";
  }

  renameFile() {
    const newFileName = this.editFileName.value.trim();

    if (!newFileName || !this.currentEditingFile) {
      return;
    }

    if (newFileName === this.currentEditingFile) {
      // No change, just close
      this.closeEditModal();
      return;
    }

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log("❌ WebSocket not connected", "system");
      return;
    }

    this.showLoadingScreen("Renaming file...");

    const message = {
      type: "file.rename",
      data: {
        filePath: this.audioInput.value, // Full path to current file
        newName: newFileName, // Just the new filename
      },
      id: this.generateUUID(),
    };

    if (this.isDebugMode) {
      this.log("📤 SENDING: " + JSON.stringify(message, null, 2), "send");
    }
    this.ws.send(JSON.stringify(message));

    // Don't close modal yet - wait for response
  }

  deleteFile() {
    if (!this.currentEditingFile) {
      return;
    }

    if (
      !confirm(`Are you sure you want to delete ${this.currentEditingFile}?`)
    ) {
      return;
    }

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log("❌ WebSocket not connected", "system");
      return;
    }

    this.showLoadingScreen("Deleting file...");

    // Send websocket message for file delete
    const message = {
      type: "file.delete",
      data: {
        filePath: this.audioInput.value, // Full path to current file
      },
      id: this.generateUUID(),
    };

    if (this.isDebugMode) {
      this.log("📤 SENDING: " + JSON.stringify(message, null, 2), "send");
    }
    this.ws.send(JSON.stringify(message));

    // Don't close modal yet - wait for response
  }

  onFileRenameSuccess(data) {
    this.hideLoadingScreen();
    this.log(
      `✅ File renamed from ${data.fileName} to ${data.newName}`,
      "system"
    );
    this.closeEditModal();

    // Refresh the dropdown and select the renamed file
    this.loadAudioFiles().then(() => {
      const fileDir = data.fileName.split("/").slice(0, -1).join("/");
      this.audioInput.value = `${fileDir}/${data.newName}`;
      this.validateForm();
      this.saveState();
    });
  }

  onFileRenameError(data) {
    this.hideLoadingScreen();
    this.log(`❌ Failed to rename file: ${data.message}`, "system");
  }

  onFileDeleteSuccess(data) {
    this.hideLoadingScreen();
    this.log(`✅ File deleted: ${data.fileName}`, "system");
    this.closeEditModal();

    // Refresh the dropdown
    this.loadAudioFiles();
  }

  onFileDeleteError(data) {
    this.hideLoadingScreen();
    this.log(`❌ Failed to delete file: ${data.message}`, "system");
  }

  // Image file modal functionality
  async openImageEditModal() {
    const selectedValue = this.pictureFileInput.value;
    if (!selectedValue || selectedValue === "") {
      return;
    }

    // Extract filename from server path
    const serverImagePath = window.PIrateRFConfig.serverPaths.imageUploads;
    if (!selectedValue.startsWith(serverImagePath)) {
      return;
    }

    // Store only the filename for display (same pattern as audio)
    this.currentEditingImageFile = selectedValue.replace(
      `${serverImagePath}/`,
      ""
    );

    // Set the filename in the input
    this.editImageName.value = this.currentEditingImageFile;

    // Show the modal
    this.imageEditModal.style.display = "flex";
  }

  closeImageEditModal() {
    this.imageEditModal.style.display = "none";
    this.currentEditingImageFile = null;
    this.editImageName.value = "";
  }

  renameImageFile() {
    const newFileName = this.editImageName.value.trim();

    if (!newFileName || !this.currentEditingImageFile) {
      return;
    }

    if (newFileName === this.currentEditingImageFile) {
      // No change, just close
      this.closeImageEditModal();
      return;
    }

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log("❌ WebSocket not connected", "system");
      return;
    }

    const message = {
      type: "file.rename",
      data: {
        filePath: this.pictureFileInput.value, // Full path to current file (same as audio pattern)
        newName: newFileName, // Just the new filename
      },
    };

    this.ws.send(JSON.stringify(message));
    this.showLoadingScreen();

    // Don't close modal yet - wait for response
  }

  deleteImageFile() {
    if (!this.currentEditingImageFile) {
      return;
    }

    if (
      !confirm(
        `Are you sure you want to delete "${this.currentEditingImageFile}"?`
      )
    ) {
      return;
    }

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log("❌ WebSocket not connected", "system");
      return;
    }

    const message = {
      type: "file.delete",
      data: {
        filePath: this.pictureFileInput.value, // Full path to file (same as audio pattern)
      },
    };

    this.ws.send(JSON.stringify(message));
    this.showLoadingScreen();

    // Don't close modal yet - wait for response
  }

  onImageFileRenameSuccess(data) {
    this.hideLoadingScreen();
    this.log(
      `✅ Image file renamed from ${data.fileName} to ${data.newName}`,
      "system"
    );
    this.closeImageEditModal();

    // Refresh the dropdown and select the renamed file
    this.loadImageFiles().then(() => {
      const fileDir = data.fileName.split("/").slice(0, -1).join("/");
      this.pictureFileInput.value = `${fileDir}/${data.newName}`;
      this.validateForm();
      this.saveState();
    });
  }

  onImageFileRenameError(data) {
    this.hideLoadingScreen();
    this.log(`❌ Failed to rename image file: ${data.message}`, "system");
  }

  onImageFileDeleteSuccess(data) {
    this.hideLoadingScreen();
    this.log(`✅ Image file deleted: ${data.fileName}`, "system");
    this.closeImageEditModal();
    this.loadImageFiles();
  }

  onImageFileDeleteError(data) {
    this.hideLoadingScreen();
    this.log(`❌ Failed to delete image file: ${data.message}`, "system");
  }

  // Playlist functionality
  async openPlaylistModal() {
    await this.loadPlaylistFiles();
    this.validatePlaylistCreation(); // Set initial button state
    this.playlistModal.style.display = "flex";
  }

  closePlaylistModal() {
    this.playlistModal.style.display = "none";
  }

  async loadPlaylistFiles() {
    try {
      // Load SFX files
      const sfxResponse = await this.customFetch(
        window.PIrateRFConfig.paths.audioSFXFiles,
        {},
        "Loading audio files..."
      );
      if (sfxResponse.ok) {
        const sfxFiles = await sfxResponse.json();
        this.renderFileList(
          this.sfxFileList,
          sfxFiles.filter((f) => !f.isDir && f.name.endsWith(".wav")),
          "sfx"
        );
      }

      // Load uploaded files
      const uploadResponse = await this.customFetch(
        window.PIrateRFConfig.paths.audioUploadFiles,
        {},
        "Loading audio files..."
      );
      if (uploadResponse.ok) {
        const uploadFiles = await uploadResponse.json();
        this.renderFileList(
          this.uploadedFileList,
          uploadFiles.filter((f) => !f.isDir && f.name.endsWith(".wav")),
          "uploads"
        );
      }
    } catch (error) {
      this.log(`❌ Failed to load playlist files: ${error.message}`, "system");
    }
  }

  renderFileList(container, files, type) {
    container.innerHTML = "";

    if (files.length === 0) {
      container.innerHTML =
        '<div style="color: #666; padding: 10px;">No files available</div>';
      return;
    }

    files.sort((a, b) => a.name.localeCompare(b.name));

    files.forEach((file) => {
      const fileItem = document.createElement("div");
      fileItem.className = "file-item";
      fileItem.innerHTML = `
        <span class="file-name" title="${file.name}">${file.name}</span>
        <div class="file-buttons">
          <button class="play-btn" data-file="${file.name}" data-type="${type}" title="Play audio">▶️</button>
          <button class="add-btn" data-file="${file.name}" data-type="${type}">Add</button>
        </div>
      `;

      const playBtn = fileItem.querySelector(".play-btn");
      const addBtn = fileItem.querySelector(".add-btn");

      playBtn.addEventListener("click", () =>
        this.toggleAudioPlayback(file.name, type, playBtn)
      );
      addBtn.addEventListener("click", () =>
        this.addToPlaylist(file.name, type)
      );

      container.appendChild(fileItem);
    });
  }

  addToPlaylist(fileName, type) {
    const displayPath = this.buildFilePath(fileName, type); // HTTP path for UI
    const serverPath = this.buildFilePath(fileName, type, true); // Server path for backend
    const justFilename = fileName.includes("/")
      ? fileName.split("/").pop()
      : fileName;

    this.playlist.push({
      name: justFilename,
      path: serverPath,
      displayPath: displayPath,
      type: type,
    });
    this.renderPlaylist();
    if (this.isDebugMode) {
      this.log(`📋 Added to playlist: ${fileName}`, "system");
    }
  }

  renderPlaylist() {
    if (this.playlist.length === 0) {
      this.playlistItems.innerHTML =
        '<div class="empty-playlist">Click "Add" next to files to build your playlist</div>';
    } else {
      this.playlistItems.innerHTML = "";
    }

    this.validatePlaylistCreation();

    this.playlist.forEach((item, index) => {
      const playlistItem = document.createElement("div");
      playlistItem.className = "playlist-item";
      playlistItem.innerHTML = `
        <span class="playlist-item-number">${index + 1}.</span>
        <span class="playlist-item-name" title="${item.name}">${
        item.name
      }</span>
        <button class="remove-btn" data-index="${index}">Remove</button>
      `;

      const removeBtn = playlistItem.querySelector(".remove-btn");
      removeBtn.addEventListener("click", () => this.removeFromPlaylist(index));

      this.playlistItems.appendChild(playlistItem);
    });
  }

  removeFromPlaylist(index) {
    const removedItem = this.playlist.splice(index, 1)[0];
    this.renderPlaylist();
    if (this.isDebugMode) {
      this.log(`📋 Removed from playlist: ${removedItem.name}`, "system");
    }
  }

  clearPlaylist() {
    this.playlist = [];
    this.renderPlaylist();
    if (this.isDebugMode) {
      this.log(`📋 Playlist cleared`, "system");
    }
  }

  validatePlaylistCreation() {
    const playlistName = this.playlistName.value.trim();
    const hasPlaylistItems = this.playlist.length > 0;
    const hasValidName = playlistName.length > 0;

    this.createPlaylistBtn.disabled = !(hasPlaylistItems && hasValidName);
  }

  createPlaylist() {
    const playlistName = this.playlistName.value.trim();

    if (!playlistName) {
      this.log("❌ Please enter a playlist name", "system");
      return;
    }

    if (this.playlist.length === 0) {
      this.log("❌ Playlist is empty. Add some files first.", "system");
      return;
    }

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log("❌ WebSocket not connected", "system");
      return;
    }

    this.showLoadingScreen("Creating playlist...");

    const message = {
      type: "audio.playlist.create",
      data: {
        playlistFileName: playlistName,
        files: this.playlist.map((item) => item.path),
      },
      id: this.generateUUID(),
    };

    if (this.isDebugMode) {
      this.log("📤 SENDING: " + JSON.stringify(message, null, 2), "send");
    }
    this.ws.send(JSON.stringify(message));

    // Clear any previous error
    this.playlistError.style.display = "none";

    this.log(
      `🎵 Creating playlist: ${playlistName} with ${this.playlist.length} files`,
      "system"
    );
  }

  onPlaylistCreateSuccess(data) {
    this.hideLoadingScreen();
    this.log(
      `✅ Playlist created successfully: ${data.playlistName}`,
      "system"
    );
    this.closePlaylistModal();

    // Clear the playlist
    this.playlist = [];
    this.playlistName.value = "";

    // Refresh the dropdown and select the newest file (the new playlist)
    this.loadAudioFiles();
  }

  onPlaylistCreateError(data) {
    this.hideLoadingScreen();
    this.log(`❌ Failed to create playlist: ${data.message}`, "system");

    // Show error in the playlist modal
    this.playlistError.textContent = `❌ Error: ${data.message}`;
    this.playlistError.style.display = "block";

    // Hide error after 5 seconds
    setTimeout(() => {
      this.playlistError.style.display = "none";
    }, 5000);
  }

  // Audio playback controls
  playSelectedAudio() {
    const selectedPath = this.audioInput.value;
    if (!selectedPath) return;

    this.toggleAudioPlayback(selectedPath, null, this.playAudioBtn);
  }

  toggleAudioPlayback(fileName, type, buttonElement) {
    const audioPath = this.buildFilePath(fileName, type);

    // Stop current audio if different file
    if (this.currentAudioPath && this.currentAudioPath !== audioPath) {
      this.stopAudio();
    }

    if (this.audioPlayer.paused || this.currentAudioPath !== audioPath) {
      this.playAudio(audioPath, buttonElement);
    } else {
      this.stopAudio();
    }
  }

  playAudio(audioPath, buttonElement) {
    this.currentAudioPath = audioPath;
    this.currentPlayButton = buttonElement;

    this.audioPlayer.src = audioPath;
    this.audioPlayer
      .play()
      .then(() => {
        if (buttonElement) {
          buttonElement.textContent = "⏹️";
          buttonElement.title = "Stop audio";
        }
        if (this.isDebugMode) {
          this.log(`🔊 Playing: ${audioPath.split("/").pop()}`, "system");
        }
      })
      .catch((error) => {
        this.log(`❌ Failed to play audio: ${error.message}`, "system");
        this.stopAudio();
      });

    // Auto-stop when audio ends
    this.audioPlayer.onended = () => this.stopAudio();
  }

  stopAudio() {
    this.audioPlayer.pause();
    this.audioPlayer.currentTime = 0;

    if (this.currentPlayButton) {
      this.currentPlayButton.textContent = "▶️";
      this.currentPlayButton.title =
        this.currentPlayButton === this.playAudioBtn
          ? "Play selected audio"
          : "Play audio";
    }

    // Reset all play buttons in file lists
    document.querySelectorAll(".play-btn").forEach((btn) => {
      if (btn !== this.playAudioBtn) {
        btn.textContent = "▶️";
        btn.title = "Play audio";
      }
    });

    // Reset intro/outro buttons
    if (this.playIntroBtn) {
      this.playIntroBtn.textContent = "▶️";
      this.playIntroBtn.title = "Play intro";
    }
    if (this.playOutroBtn) {
      this.playOutroBtn.textContent = "▶️";
      this.playOutroBtn.title = "Play outro";
    }

    this.currentAudioPath = null;
    this.currentPlayButton = null;
  }

  togglePlayMode() {
    if (this.playModeToggle.classList.contains("play-once")) {
      // Switch to continuous mode
      this.playModeToggle.classList.remove("play-once");
      this.playModeToggle.classList.add("continuous");
      this.playModeToggle.textContent = "🔁";
      this.playModeToggle.title = "Play continuously";
    } else {
      // Switch to play once mode
      this.playModeToggle.classList.add("play-once");
      this.playModeToggle.classList.remove("continuous");
      this.playModeToggle.textContent = "⏭️";
      this.playModeToggle.title = "Play once";
    }
  }

  toggleIntroOutro() {
    if (this.introOutroControls.classList.contains("hidden")) {
      // Show intro/outro controls
      this.introOutroControls.classList.remove("hidden");
      this.introOutroToggle.classList.add("active");
      this.loadSfxFiles(); // Load SFX files when opening
    } else {
      // Hide intro/outro controls
      this.introOutroControls.classList.add("hidden");
      this.introOutroToggle.classList.remove("active");
    }
  }

  async loadSfxFiles() {
    try {
      // Use existing SFX loading from playlist functionality
      const sfxResponse = await fetch(
        window.PIrateRFConfig.paths.audioSFXFiles
      );
      if (sfxResponse.ok) {
        const sfxFiles = await sfxResponse.json();
        this.populateSfxDropdowns(
          sfxFiles.filter((f) => !f.isDir && f.name.endsWith(".wav"))
        );
      }
    } catch (error) {
      this.log(`❌ Failed to load SFX files: ${error.message}`, "system");
    }
  }

  populateSfxDropdowns(sfxFiles) {
    // Clear existing options (keep "No intro/outro")
    this.introSelect.innerHTML = '<option value="">No intro</option>';
    this.outroSelect.innerHTML = '<option value="">No outro</option>';

    // Add SFX files to both dropdowns
    sfxFiles.forEach((file) => {
      const serverPath = this.buildFilePath(file.name, "sfx", true); // Server path for backend
      const justFilename = file.name.includes("/")
        ? file.name.split("/").pop()
        : file.name;
      const introOption = new Option(justFilename, serverPath);
      const outroOption = new Option(justFilename, serverPath);
      this.introSelect.add(introOption);
      this.outroSelect.add(outroOption);
    });

    // Restore saved intro/outro selections if they still exist
    this.restoreSfxSelections();
  }

  onIntroOutroChange() {
    // Enable/disable play buttons based on selection
    this.playIntroBtn.disabled = !this.introSelect.value;
    this.playOutroBtn.disabled = !this.outroSelect.value;
  }

  playIntroOutro(type) {
    const fileName =
      type === "intro" ? this.introSelect.value : this.outroSelect.value;
    if (!fileName) return;

    const buttonElement =
      type === "intro" ? this.playIntroBtn : this.playOutroBtn;

    // If currently playing this audio, stop it
    if (buttonElement.textContent === "⏹️") {
      this.stopAudio();
      return;
    }

    // Construct proper SFX file path and play
    const fullPath = this.buildFilePath(fileName, "sfx");
    this.playAudio(fullPath, buttonElement);
  }

  generateUUID() {
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(
      /[xy]/g,
      function (c) {
        const r = (Math.random() * 16) | 0;
        const v = c == "x" ? r : (r & 0x3) | 0x8;
        return v.toString(16);
      }
    );
  }

  // Custom fetch with loading screen
  async customFetch(url, options = {}, loadingMessage = "Loading...") {
    this.showLoadingScreen(loadingMessage);

    try {
      const response = await fetch(url, options);
      return response;
    } finally {
      this.hideLoadingScreen();
    }
  }

  // Loading screen methods
  showLoadingScreen(message = "Loading...") {
    this.loadingText.textContent = message;
    this.loadingOverlay.style.display = "flex";
  }

  hideLoadingScreen() {
    this.loadingOverlay.style.display = "none";
  }

  // Error notification methods
  showErrorNotification(message, errorType = null) {
    // Clear any existing timeout
    if (this.errorTimeout) {
      clearTimeout(this.errorTimeout);
    }

    // Set the error message
    this.errorMessage.textContent = message;

    // Set error type data attribute for tracking
    if (errorType) {
      this.errorNotification.setAttribute("data-error-type", errorType);
    } else {
      this.errorNotification.removeAttribute("data-error-type");
    }

    // Show the notification
    this.errorNotification.style.display = "block";

    // Auto-hide after 10 seconds
    this.errorTimeout = setTimeout(() => {
      this.hideErrorNotification();
    }, 10000);
  }

  hideErrorNotification() {
    this.errorNotification.style.display = "none";

    // Clear timeout if it exists
    if (this.errorTimeout) {
      clearTimeout(this.errorTimeout);
      this.errorTimeout = null;
    }
  }

  clearWebSocketErrors() {
    // Hide WebSocket-related error notifications
    const errorType = this.errorNotification.getAttribute("data-error-type");
    if (errorType === "websocket" || errorType === "websocket-disconnected") {
      this.hideErrorNotification();
    }
  }

  // Select saved audio file if it exists, otherwise select first file
  selectSavedOrFirstAudioFile() {
    try {
      const savedState = localStorage.getItem("piraterf_state");
      if (savedState) {
        const state = JSON.parse(savedState);
        if (state.audio) {
          // Check if the saved audio file still exists in the dropdown
          const option = Array.from(this.audioInput.options).find(
            (opt) => opt.value === state.audio
          );
          if (option) {
            this.audioInput.value = state.audio;
            return; // Found and selected saved file
          }
        }
      }
    } catch (e) {
      console.warn("Failed to check saved audio file:", e);
    }

    // Fallback: select first (newest) file if no saved selection or saved file doesn't exist
    if (this.audioInput.options.length > 0) {
      this.audioInput.selectedIndex = 0;
    }
  }

  // Restore saved intro/outro selections if they still exist in dropdowns
  restoreSfxSelections() {
    try {
      // Use the current state object instead of parsing localStorage again

      // Restore intro selection if it exists
      if (this.state.pifmrds && this.state.pifmrds.introSelect) {
        const introOption = Array.from(this.introSelect.options).find(
          (opt) => opt.value === this.state.pifmrds.introSelect
        );
        if (introOption) {
          this.introSelect.value = this.state.pifmrds.introSelect;
        }
      }

      // Restore outro selection if it exists
      if (this.state.pifmrds && this.state.pifmrds.outroSelect) {
        const outroOption = Array.from(this.outroSelect.options).find(
          (opt) => opt.value === this.state.pifmrds.outroSelect
        );
        if (outroOption) {
          this.outroSelect.value = this.state.pifmrds.outroSelect;
        }
      }

      // Update play button states
      this.onIntroOutroChange();
    } catch (e) {
      console.warn("Failed to restore SFX selections:", e);
    }
  }

  // Update state object from DOM and save to localStorage
  saveState() {
    // Ensure state object structure exists
    if (!this.state.pifmrds) this.state.pifmrds = {};
    if (!this.state.morse) this.state.morse = {};
    if (!this.state.tune) this.state.tune = {};
    if (!this.state.spectrumpaint) this.state.spectrumpaint = {};
    if (!this.state.pichirp) this.state.pichirp = {};

    // Update state object from current DOM values
    this.state.modulename = this.moduleSelect.value;

    // Update PIFMRDS state
    this.state.pifmrds.freq = this.freqInput.value;
    this.state.pifmrds.audio = this.audioInput.value;
    this.state.pifmrds.pi = this.piInput.value;
    this.state.pifmrds.ps = this.psInput.value;
    this.state.pifmrds.rt = this.rtInput.value;
    this.state.pifmrds.timeout = document.getElementById("timeout").value;
    this.state.pifmrds.playOnce =
      this.playModeToggle.classList.contains("play-once");
    this.state.pifmrds.introOutroToggled =
      this.introOutroToggle.classList.contains("active");
    this.state.pifmrds.introSelect = this.introSelect.value;
    this.state.pifmrds.outroSelect = this.outroSelect.value;

    // Update MORSE state
    this.state.morse.freq = document.getElementById("morseFreq")?.value || "";
    this.state.morse.rate = document.getElementById("morseRate")?.value || "";
    this.state.morse.message =
      document.getElementById("morseMessage")?.value || "";

    // Update TUNE state
    this.state.tune.freq = document.getElementById("tuneFreq")?.value || "";
    this.state.tune.exitImmediate =
      document.getElementById("tuneExitImmediate")?.checked || false;
    this.state.tune.ppm = document.getElementById("tunePPM")?.value || "";

    // Update SPECTRUMPAINT state
    this.state.spectrumpaint.frequency =
      this.spectrumpaintFreqInput?.value || "";
    this.state.spectrumpaint.pictureFile = this.pictureFileInput?.value || "";
    this.state.spectrumpaint.excursion = this.excursionInput?.value || "";

    // Update PICHIRP state
    this.state.pichirp.frequency = this.pichirpFreqInput?.value || "";
    this.state.pichirp.bandwidth = this.pichirpBandwidthInput?.value || "";
    this.state.pichirp.time = this.pichirpTimeInput?.value || "";

    try {
      localStorage.setItem("piraterf_state", JSON.stringify(this.state));
    } catch (e) {
      console.warn("Failed to save state to localStorage:", e);
    }
  }

  // Load state from localStorage and sync to DOM
  restoreState() {
    try {
      const savedState = localStorage.getItem("piraterf_state");
      if (savedState) {
        const parsedState = JSON.parse(savedState);
        // Merge with default state to ensure all properties exist
        this.state = {
          ...this.state,
          ...parsedState,
          pifmrds: { ...this.state.pifmrds, ...parsedState.pifmrds },
          morse: { ...this.state.morse, ...parsedState.morse },
          tune: { ...this.state.tune, ...parsedState.tune },
          spectrumpaint: {
            ...this.state.spectrumpaint,
            ...parsedState.spectrumpaint,
          },
          pichirp: { ...this.state.pichirp, ...parsedState.pichirp },
        };
      }

      // Sync state to DOM elements
      this.syncStateToDOM();
    } catch (e) {
      console.warn("Failed to restore state from localStorage:", e);
    }
  }

  // Sync the current state object to DOM elements
  syncStateToDOM() {
    // Ensure state object structure exists
    if (!this.state.pifmrds) this.state.pifmrds = {};
    if (!this.state.morse) this.state.morse = {};
    if (!this.state.tune) this.state.tune = {};
    if (!this.state.spectrumpaint) this.state.spectrumpaint = {};
    if (!this.state.pichirp) this.state.pichirp = {};

    // Sync module selection
    if (this.state.modulename) this.moduleSelect.value = this.state.modulename;

    // Sync PIFMRDS form inputs
    if (this.state.pifmrds.freq) this.freqInput.value = this.state.pifmrds.freq;
    if (this.state.pifmrds.audio)
      this.audioInput.value = this.state.pifmrds.audio;
    if (this.state.pifmrds.pi) this.piInput.value = this.state.pifmrds.pi;
    if (this.state.pifmrds.ps) this.psInput.value = this.state.pifmrds.ps;
    if (this.state.pifmrds.rt) this.rtInput.value = this.state.pifmrds.rt;
    if (this.state.pifmrds.timeout)
      document.getElementById("timeout").value = this.state.pifmrds.timeout;

    // Sync play mode toggle (PIFMRDS only)
    if (this.state.pifmrds.playOnce !== undefined) {
      if (this.state.pifmrds.playOnce) {
        this.playModeToggle.classList.add("play-once");
        this.playModeToggle.classList.remove("continuous");
        this.playModeToggle.textContent = "⏭️";
        this.playModeToggle.title = "Play once";
      } else {
        this.playModeToggle.classList.add("continuous");
        this.playModeToggle.classList.remove("play-once");
        this.playModeToggle.textContent = "🔁";
        this.playModeToggle.title = "Play continuously";
      }
    }

    // Sync intro/outro toggle (PIFMRDS only)
    if (this.state.pifmrds.introOutroToggled !== undefined) {
      if (this.state.pifmrds.introOutroToggled) {
        this.introOutroToggle.classList.add("active");
        this.introOutroControls.classList.remove("hidden");
      } else {
        this.introOutroToggle.classList.remove("active");
        this.introOutroControls.classList.add("hidden");
      }
      // Always load SFX files to populate dropdowns for state restoration
      this.loadSfxFiles();
    }

    // Sync MORSE form inputs
    if (this.state.morse.freq)
      document.getElementById("morseFreq").value = this.state.morse.freq;
    if (this.state.morse.rate)
      document.getElementById("morseRate").value = this.state.morse.rate;
    if (this.state.morse.message)
      document.getElementById("morseMessage").value = this.state.morse.message;

    // Sync TUNE form inputs
    if (this.state.tune.freq)
      document.getElementById("tuneFreq").value = this.state.tune.freq;
    if (this.state.tune.exitImmediate !== undefined)
      document.getElementById("tuneExitImmediate").checked =
        this.state.tune.exitImmediate;
    if (this.state.tune.ppm)
      document.getElementById("tunePPM").value = this.state.tune.ppm;

    // Sync SPECTRUMPAINT form inputs
    if (this.state.spectrumpaint.frequency && this.spectrumpaintFreqInput)
      this.spectrumpaintFreqInput.value = this.state.spectrumpaint.frequency;
    if (this.state.spectrumpaint.pictureFile && this.pictureFileInput)
      this.pictureFileInput.value = this.state.spectrumpaint.pictureFile;
    if (this.state.spectrumpaint.excursion && this.excursionInput)
      this.excursionInput.value = this.state.spectrumpaint.excursion;

    // Sync PICHIRP form inputs
    if (this.state.pichirp.frequency && this.pichirpFreqInput)
      this.pichirpFreqInput.value = this.state.pichirp.frequency;
    if (this.state.pichirp.bandwidth && this.pichirpBandwidthInput)
      this.pichirpBandwidthInput.value = this.state.pichirp.bandwidth;
    if (this.state.pichirp.time && this.pichirpTimeInput)
      this.pichirpTimeInput.value = this.state.pichirp.time;

    // Note: intro/outro selections are restored by restoreSfxSelections() when SFX files are loaded

    // Trigger module change to show/hide appropriate form fields
    this.onModuleChange();
  }
}

// Initialize the application
document.addEventListener("DOMContentLoaded", () => {
  new PIrateRFController();
});
