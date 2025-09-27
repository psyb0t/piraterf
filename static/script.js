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
        freq: "87.9",
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
        freq: "7058000",
        rate: "20",
        message: "HACK THE PLANET",
      },

      tune: {
        freq: "144500000",
        exitImmediate: false,
        ppm: "",
      },

      spectrumpaint: {
        frequency: "144500000",
        pictureFile: "",
        excursion: "50000",
      },

      pichirp: {
        frequency: "144500000",
        bandwidth: "1000000",
        time: "5.0",
      },

      pocsag: {
        frequency: "152000000",
        baudRate: "1200",
        functionBits: "3",
        numericMode: false,
        repeatCount: "4",
        invertPolarity: false,
        debug: false,
        messages: [
          {
            address: "123456",
            message: "TEST MESSAGE",
            functionBits: "",
          },
        ],
      },

      pift8: {
        frequency: "14074000",
        message: "",
        ppm: "",
        offset: "",
        slot: "0",
        repeat: false,
      },

      pisstv: {
        frequency: "14233000",
        pictureFile: "",
      },

      pirtty: {
        frequency: "14075000",
        spaceFrequency: "",
        message: "",
      },

      fsk: {
        frequency: "144500000",
        inputType: "text",
        text: "",
        file: "",
        baudRate: "",
      },

      "audiosock-broadcast": {
        frequency: "27225000",
        sampleRate: "",
        bufferSize: "4096",
        modulation: "FM",
        gain: "1.0",
      },
    };

    this.initializeElements();
    this.isDebugMode = this.checkDebugMode();
    this.bindEvents();
    // Restore module selection immediately to show correct form
    this.restoreModuleSelection();
    this.loadAudioFiles(false).then(() => {
      // Restore full state after audio files are loaded
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
    } else if (type === "data") {
      if (forServer) {
        return `${window.PIrateRFConfig.serverPaths.dataUploads}/${justFilename}`;
      }
      return `${window.PIrateRFConfig.paths.files}/${window.PIrateRFConfig.directories.dataUploads}/${justFilename}`;
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

  debug(...args) {
    if (this.isDebugMode) {
      console.log("🐛", ...args);
    }
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
    this.pocsagForm = document.getElementById("pocsagForm");
    this.pift8Form = document.getElementById("pift8Form");
    this.pisstvForm = document.getElementById("pisstvForm");
    this.pirttyForm = document.getElementById("pirttyForm");
    this.fskForm = document.getElementById("fskForm");
    this.audioSockBroadcastForm = document.getElementById("audiosock-broadcastForm");

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

    // POCSAG form inputs
    this.pocsagFreqInput = document.getElementById("pocsagFreq");
    this.pocsagBaudRateInput = document.getElementById("pocsagBaudRate");
    this.pocsagFunctionBitsInput =
      document.getElementById("pocsagFunctionBits");
    this.pocsagRepeatCountInput = document.getElementById("pocsagRepeatCount");
    this.pocsagNumericModeInput = document.getElementById("pocsagNumericMode");
    this.pocsagInvertPolarityInput = document.getElementById(
      "pocsagInvertPolarity"
    );
    this.pocsagDebugInput = document.getElementById("pocsagDebug");
    this.pocsagMessagesContainer = document.getElementById("pocsagMessages");
    this.addMessageBtn = document.getElementById("addMessageBtn");

    // FT8 form inputs
    this.ft8FreqInput = document.getElementById("ft8Freq");
    this.ft8MessageInput = document.getElementById("ft8Message");
    this.ft8PPMInput = document.getElementById("ft8PPM");
    this.ft8OffsetInput = document.getElementById("ft8Offset");
    this.ft8SlotInput = document.getElementById("ft8Slot");
    this.ft8RepeatInput = document.getElementById("ft8Repeat");

    // PISSTV form inputs
    this.pisstvFreqInput = document.getElementById("pisstvFreq");
    this.pisstvPictureFileInput = document.getElementById("pisstvPictureFile");

    // PIRTTY form inputs
    this.pirttyFreqInput = document.getElementById("pirttyFreq");
    this.pirttySpaceFreqInput = document.getElementById("pirttySpaceFreq");
    this.pirttyMessageInput = document.getElementById("pirttyMessage");

    // FSK form inputs
    this.fskFreqInput = document.getElementById("fskFreq");
    this.fskInputTypeText = document.getElementById("fskInputTypeText");
    this.fskInputTypeFile = document.getElementById("fskInputTypeFile");
    this.fskTextInput = document.getElementById("fskText");
    this.fskFileInput = document.getElementById("fskFile");
    this.fskBaudRateInput = document.getElementById("fskBaudRate");
    this.fskTextGroup = document.getElementById("fskTextGroup");
    this.fskFileGroup = document.getElementById("fskFileGroup");
    this.refreshFskFileBtn = document.getElementById("refreshFskFileBtn");
    this.fskFileSelectBtn = document.getElementById("fskFileSelectBtn");
    this.editFskFileBtn = document.getElementById("editFskFileBtn");

    // AudioSock Broadcast form inputs
    this.audioSockBroadcastFreqInput = document.getElementById("audioSockBroadcastFreq");
    this.audioSockBroadcastSampleRateInput = document.getElementById("audioSockBroadcastSampleRate");
    this.audioSockBroadcastBufferSizeInput = document.getElementById("audioSockBroadcastBufferSize");
    this.audioSockBroadcastModulationInput = document.getElementById("audioSockBroadcastModulation");
    this.audioSockBroadcastGainInput = document.getElementById("audioSockBroadcastGain");
    this.fskDataFile = document.getElementById("fskDataFile");

    // PISSTV image control buttons
    this.refreshPisstvImageBtn = document.getElementById("refreshPisstvImageBtn");
    this.editPisstvImageBtn = document.getElementById("editPisstvImageBtn");
    this.pisstvImageSelectBtn = document.getElementById("pisstvImageSelectBtn");

    this.refreshImageBtn = document.getElementById("refreshImageBtn");
    this.editImageBtn = document.getElementById("editImageBtn");
    this.imageSelectBtn = document.getElementById("imageSelectBtn");
    this.imageFile = document.getElementById("imageFile");
    this.imageUploadStatus = document.getElementById("imageUploadStatus");

    // Audio file dropdown
    this.refreshAudioBtn = document.getElementById("refreshAudioBtn");
    this.editAudioBtn = document.getElementById("editAudioBtn");

    // Modal elements (unified for all file types)
    this.fileEditModal = document.getElementById("fileEditModal");
    this.fileEditModalTitle = document.getElementById("fileEditModalTitle");
    this.modalCloseBtn = document.getElementById("modalCloseBtn");
    this.modalCancelBtn = document.getElementById("modalCancelBtn");
    this.renameFileBtn = document.getElementById("renameFileBtn");
    this.editFileName = document.getElementById("editFileName");
    this.deleteFileBtn = document.getElementById("deleteFileBtn");

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
    this.currentEditFile = null;
    this.currentEditDirectory = null;
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
      if (!this.isRestoring) {
        this.saveState();
      }
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
    this.editAudioBtn.addEventListener("click", () => this.openAudioEditModal());
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

    // Modal click handlers removed - modals only close via buttons

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
    this.editImageBtn.addEventListener("click", () => {
      const selectedFile = this.pictureFileInput.value;
      if (selectedFile) {
        const fileName = selectedFile.split('/').pop();
        this.openFileEditModal("image", fileName, "imageUploads");
      } else {
        this.log("❌ No spectrum paint image file selected", "system");
      }
    });
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

    // POCSAG module form events
    this.pocsagFreqInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.pocsagBaudRateInput.addEventListener("change", () => {
      this.saveState();
      this.validateForm();
    });
    this.pocsagFunctionBitsInput.addEventListener("change", () => {
      this.saveState();
      this.validateForm();
    });
    this.pocsagRepeatCountInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.pocsagNumericModeInput.addEventListener("click", () => {
      this.togglePOCSAGOption(this.pocsagNumericModeInput);
      this.saveState();
      this.validateForm();
    });
    this.pocsagInvertPolarityInput.addEventListener("click", () => {
      this.togglePOCSAGOption(this.pocsagInvertPolarityInput);
      this.saveState();
      this.validateForm();
    });
    this.pocsagDebugInput.addEventListener("click", () => {
      this.togglePOCSAGOption(this.pocsagDebugInput);
      this.saveState();
      this.validateForm();
    });

    // FT8 module form events
    this.debug("🔧 Setting up FT8 event listeners - freq element:", this.ft8FreqInput, "msg element:", this.ft8MessageInput);
    this.ft8FreqInput.addEventListener("input", () => {
      this.debug("🔧 FT8 freq changed, fucking finally:", this.ft8FreqInput.value);
      this.saveState();
      this.validateForm();
    });
    this.ft8MessageInput.addEventListener("input", () => {
      this.debug("💬 FT8 message changed, you bastard:", this.ft8MessageInput.value);
      this.saveState();
      this.validateForm();
    });
    this.ft8PPMInput.addEventListener("input", () => {
      this.debug("⚡ FT8 ppm tweaked:", this.ft8PPMInput.value);
      this.saveState();
      this.validateForm();
    });
    this.ft8OffsetInput.addEventListener("input", () => {
      this.debug("📡 FT8 offset hacked:", this.ft8OffsetInput.value);
      this.saveState();
      this.validateForm();
    });
    this.ft8SlotInput.addEventListener("change", () => {
      this.debug("🎰 FT8 slot switched:", this.ft8SlotInput.value);
      this.saveState();
      this.validateForm();
    });
    this.ft8RepeatInput.addEventListener("click", () => {
      this.ft8RepeatInput.classList.toggle("active");
      this.debug("🔁 FT8 repeat toggled, shit works:", this.ft8RepeatInput.classList.contains("active"));
      this.saveState();
      this.validateForm();
    });

    // PISSTV module form events
    this.pisstvFreqInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.pisstvPictureFileInput.addEventListener("change", () => {
      this.onPisstvImageFileChange();
      this.saveState();
      this.validateForm();
    });

    // PISSTV image control buttons
    this.refreshPisstvImageBtn.addEventListener("click", () => this.loadImageFiles());
    this.editPisstvImageBtn.addEventListener("click", () => {
      const selectedFile = this.pisstvPictureFileInput.value;
      if (selectedFile) {
        const fileName = selectedFile.split('/').pop();
        this.openFileEditModal("image", fileName, "imageUploads");
      } else {
        this.log("❌ No SSTV image file selected", "system");
      }
    });
    this.pisstvImageSelectBtn.addEventListener("click", () => this.imageFile.click());

    // PIRTTY module form events
    this.pirttyFreqInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.pirttySpaceFreqInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.pirttyMessageInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });

    // FSK module form events
    this.fskFreqInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.fskInputTypeText.addEventListener("click", () => {
      this.setFskInputType("text");
    });
    this.fskInputTypeFile.addEventListener("click", () => {
      this.setFskInputType("file");
    });
    this.fskTextInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.fskFileInput.addEventListener("change", () => {
      this.onFskFileChange();
      this.saveState();
      this.validateForm();
    });
    this.fskBaudRateInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });

    // FSK file control buttons
    this.refreshFskFileBtn.addEventListener("click", () => this.loadDataFiles());
    this.editFskFileBtn.addEventListener("click", () => this.openDataFileEditModal());
    this.fskFileSelectBtn.addEventListener("click", () => this.fskDataFile.click());
    this.fskDataFile.addEventListener("change", (event) => this.handleDataFileUpload(event));
    this.fskFileInput.addEventListener("change", () => {
      this.onFskFileChange();
      this.saveState();
      this.validateForm();
    });

    // AudioSock Broadcast module form events
    this.audioSockBroadcastFreqInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.audioSockBroadcastSampleRateInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });
    this.audioSockBroadcastBufferSizeInput.addEventListener("change", () => {
      this.saveState();
      this.validateForm();
    });
    this.audioSockBroadcastModulationInput.addEventListener("change", () => {
      this.saveState();
      this.validateForm();
    });
    this.audioSockBroadcastGainInput.addEventListener("input", () => {
      this.saveState();
      this.validateForm();
    });

    // POCSAG messages management
    this.addMessageBtn.addEventListener("click", () => this.addPOCSAGMessage());
    this.bindPOCSAGMessageEvents();
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
        // Check file type based on the file path
        if (
          message.data.fileName &&
          message.data.fileName.includes(`/${window.PIrateRFConfig.directories.imageUploads}/`)
        ) {
          this.onImageFileRenameSuccess(message.data);
        } else if (
          message.data.fileName &&
          message.data.fileName.includes(`/${window.PIrateRFConfig.directories.dataUploads}/`)
        ) {
          this.onDataFileRenameSuccess(message.data);
        } else {
          this.onFileRenameSuccess(message.data);
        }
        break;
      case "file.rename.error":
        // Check file type based on the file path
        if (
          message.data.fileName &&
          message.data.fileName.includes(`/${window.PIrateRFConfig.directories.imageUploads}/`)
        ) {
          this.onImageFileRenameError(message.data);
        } else if (
          message.data.fileName &&
          message.data.fileName.includes(`/${window.PIrateRFConfig.directories.dataUploads}/`)
        ) {
          this.onDataFileRenameError(message.data);
        } else {
          this.onFileRenameError(message.data);
        }
        break;
      case "file.delete.success":
        this.debug(`File delete success: ${message.data.fileName}`);

        // Check file type based on the file path
        if (
          message.data.fileName &&
          message.data.fileName.includes(`/${window.PIrateRFConfig.directories.imageUploads}/`)
        ) {
          this.debug("Calling onImageFileDeleteSuccess");
          this.onImageFileDeleteSuccess(message.data);
        } else if (
          message.data.fileName &&
          message.data.fileName.includes(`/${window.PIrateRFConfig.directories.dataUploads}/`)
        ) {
          this.debug("Calling onDataFileDeleteSuccess");
          this.onDataFileDeleteSuccess(message.data);
        } else {
          this.debug("Calling onFileDeleteSuccess");
          this.onFileDeleteSuccess(message.data);
        }
        break;
      case "file.delete.error":
        // Check file type based on the file path
        if (
          message.data.fileName &&
          message.data.fileName.includes(`/${window.PIrateRFConfig.directories.imageUploads}/`)
        ) {
          this.onImageFileDeleteError(message.data);
        } else if (
          message.data.fileName &&
          message.data.fileName.includes(`/${window.PIrateRFConfig.directories.dataUploads}/`)
        ) {
          this.onDataFileDeleteError(message.data);
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
    this.debug("🔄 Module switched to:", module);

    // Hide all module forms
    this.pifmrdsForm.classList.add("hidden");
    this.morseForm.classList.add("hidden");
    this.tuneForm.classList.add("hidden");
    this.spectrumpaintForm.classList.add("hidden");
    this.pichirpForm.classList.add("hidden");
    this.pocsagForm.classList.add("hidden");
    this.pift8Form.classList.add("hidden");
    this.pisstvForm.classList.add("hidden");
    this.pirttyForm.classList.add("hidden");
    this.fskForm.classList.add("hidden");
    this.audioSockBroadcastForm.classList.add("hidden");

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
        // Get saved filename from state and extract just the filename
        const savedSpectrumFile = this.state.spectrumpaint?.pictureFile;
        const spectrumFilename = savedSpectrumFile ? savedSpectrumFile.split('/').pop() : null;
        this.loadImageFiles(spectrumFilename);
        break;
      case "pichirp":
        this.pichirpForm.classList.remove("hidden");
        break;
      case "pocsag":
        this.pocsagForm.classList.remove("hidden");
        break;
      case "pift8":
        this.pift8Form.classList.remove("hidden");
        break;
      case "pisstv":
        this.pisstvForm.classList.remove("hidden");
        // Get saved filename from state and extract just the filename
        const savedPisstvFile = this.state.pisstv?.pictureFile;
        const pisstvFilename = savedPisstvFile ? savedPisstvFile.split('/').pop() : null;
        this.loadImageFiles(pisstvFilename);
        break;
      case "pirtty":
        this.pirttyForm.classList.remove("hidden");
        break;
      case "fsk":
        this.fskForm.classList.remove("hidden");
        this.loadDataFiles(false);
        break;
      case "audiosock-broadcast":
        this.audioSockBroadcastForm.classList.remove("hidden");
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
    this.debug("🔍 validateForm() called for module:", module);
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
      case "pocsag":
        isValid = module && this.pocsagFreqInput.value && this.validatePOCSAGMessages();
        break;
      case "pift8":
        isValid = module && this.ft8FreqInput.value && this.ft8MessageInput.value;
        this.debug("🔍 FT8 validation - module:", module, "freq:", this.ft8FreqInput.value, "msg:", this.ft8MessageInput.value, "isValid:", isValid);
        break;
      case "pisstv":
        isValid = module && this.pisstvFreqInput.value && this.pisstvPictureFileInput.value;
        break;
      case "pirtty":
        isValid = module && this.pirttyFreqInput.value && this.pirttyMessageInput.value.trim();
        break;
      case "fsk":
        const hasFreq = this.fskFreqInput.value;
        const isText = this.fskInputTypeText.classList.contains('active') && this.fskTextInput.value.trim();
        const isFile = this.fskInputTypeFile.classList.contains('active') && this.fskFileInput.value;
        isValid = module && hasFreq && (isText || isFile);
        break;
      case "audiosock-broadcast":
        // For AudioSock, socket path gets populated after connecting, so don't require it
        isValid = module && this.audioSockBroadcastFreqInput.value;
        break;
      default:
        isValid = false;
    }

    this.startBtn.disabled = !isValid || this.isExecuting;
  }

  validatePOCSAGMessages() {
    const messageElements =
      this.pocsagMessagesContainer.querySelectorAll(".pocsag-message");
    if (messageElements.length === 0) {
      return false;
    }

    for (const messageElement of messageElements) {
      const addressInput = messageElement.querySelector(
        '[data-message-field="address"]'
      );
      const messageInput = messageElement.querySelector(
        '[data-message-field="message"]'
      );

      if (!addressInput || !messageInput) {
        return false;
      }

      if (!addressInput.value.trim() || !messageInput.value.trim()) {
        return false;
      }
    }

    return true;
  }

  togglePOCSAGOption(button) {
    button.classList.toggle('active');
  }

  setFskInputType(type, skipSave = false) {
    // Update toggle button states
    this.fskInputTypeText.classList.toggle('active', type === 'text');
    this.fskInputTypeFile.classList.toggle('active', type === 'file');

    this.onFskInputTypeChange();
    if (!skipSave) {
      this.saveState();
      this.validateForm();
    }
  }

  onFskInputTypeChange() {
    const isTextMode = this.fskInputTypeText.classList.contains('active');

    // Show/hide text area vs file controls
    const textContainer = this.fskTextInput.closest('.form-group');
    const fileContainer = this.fskFileInput.closest('.form-group');
    const fileControls = document.querySelector('.fsk-file-controls');

    if (isTextMode) {
      textContainer.classList.remove('hidden');
      fileContainer.classList.add('hidden');
      if (fileControls) fileControls.classList.add('hidden');
    } else {
      textContainer.classList.add('hidden');
      fileContainer.classList.remove('hidden');
      if (fileControls) fileControls.classList.remove('hidden');
      // Load data files when switching to file mode
      this.loadDataFiles();
    }
  }

  onFskFileChange() {
    const hasSelection = this.fskFileInput.value && this.fskFileInput.value !== "";
    this.editFskFileBtn.disabled = !hasSelection;
  }

  async loadDataFiles(selectFilename = null) {
    return this.loadFiles({
      endpoint: window.PIrateRFConfig.paths.dataUploadFiles,
      selectElement: this.fskFileInput,
      fileTypes: [], // No filter, accept all files
      selectLatest: !selectFilename,
      selectFilename,
      savedStateKey: 'fsk',
      onChangeCallback: () => this.onFskFileChange(),
      noFilesText: "No data files",
      debugPrefix: "data files",
      useServerPath: true,
      pathType: "data"
    });
  }

  async handleDataFileUpload(event) {
    const file = event.target.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);
    formData.append('module', 'fsk');

    try {
      const response = await this.customFetch('/upload', {
        method: 'POST',
        body: formData
      }, "Uploading data file...");

      if (!response.ok) {
        throw new Error(`Upload failed: ${response.status}`);
      }

      const result = await response.json();
      if (result.status === 'success') {
        this.log(`✅ Uploaded: ${result.original_filename}`, "system");

        // Clear the file input after successful upload
        event.target.value = "";

        // Refresh dropdown to show new file and auto-select it (newest first)
        await this.loadDataFiles();

        this.saveState();
        this.validateForm();
      }
    } catch (error) {
      this.log(`❌ Upload error: ${error.message}`, "system");
    }

    // Clear the file input for next upload
    event.target.value = '';
  }

  openFileEditModal(fileType, selectedFile, directory) {
    if (!selectedFile) {
      this.log("❌ No file selected", "system");
      return;
    }

    // Clear previous modal state first
    this.currentEditFile = null;
    this.currentEditDirectory = null;
    this.currentFileExtension = null;
    this.editFileName.value = "";

    // Set modal title and placeholder based on file type
    const config = {
      audio: { title: "Edit Audio File", placeholder: "filename" },
      image: { title: "Edit Image File", placeholder: "filename" },
      data: { title: "Edit Data File", placeholder: "filename" }
    };

    const modalConfig = config[fileType] || { title: "Edit File", placeholder: "filename" };

    // Extract extension from the actual filename
    const lastDotIndex = selectedFile.lastIndexOf('.');
    let fileExtension = "";
    let displayName = selectedFile;

    if (lastDotIndex > 0 && lastDotIndex < selectedFile.length - 1) {
      fileExtension = selectedFile.substring(lastDotIndex);
      displayName = selectedFile.substring(0, lastDotIndex);
    }

    this.fileEditModalTitle.textContent = modalConfig.title;
    this.editFileName.placeholder = modalConfig.placeholder;
    this.editFileName.value = displayName;

    this.currentEditFile = selectedFile;
    this.currentEditDirectory = directory;
    this.currentFileExtension = fileExtension;
    this.fileEditModal.style.display = "flex";
  }

  openDataFileEditModal() {
    const selectedFile = this.fskFileInput.value;
    if (selectedFile) {
      // Extract filename with extension preserved
      const fileName = selectedFile.split('/').pop();
      this.openFileEditModal("data", fileName, "data");
    } else {
      this.log("❌ No data file selected", "system");
    }
  }


  openAudioEditModal() {
    const selectedFile = this.audioInput.value;
    if (selectedFile) {
      // Extract just the filename from the path
      const fileName = selectedFile.split('/').pop();
      this.openFileEditModal("audio", fileName, "audio");
    } else {
      this.log("❌ No audio file selected", "system");
    }
  }

  buildPOCSAGArgs() {
    const args = {};

    // Add optional frequency
    if (this.pocsagFreqInput.value.trim()) {
      args.frequency = parseFloat(this.pocsagFreqInput.value);
    }

    // Add optional baud rate
    if (this.pocsagBaudRateInput.value.trim()) {
      args.baudRate = parseInt(this.pocsagBaudRateInput.value);
    }

    // Add optional function bits
    if (this.pocsagFunctionBitsInput.value.trim()) {
      args.functionBits = parseInt(this.pocsagFunctionBitsInput.value);
    }

    // Add optional repeat count
    if (this.pocsagRepeatCountInput.value.trim()) {
      args.repeatCount = parseInt(this.pocsagRepeatCountInput.value);
    }

    // Add optional boolean flags
    if (this.pocsagNumericModeInput.classList.contains('active')) {
      args.numericMode = true;
    }

    if (this.pocsagInvertPolarityInput.classList.contains('active')) {
      args.invertPolarity = true;
    }

    if (this.pocsagDebugInput.classList.contains('active')) {
      args.debug = true;
    }

    // Build messages array
    args.messages = [];
    const messageElements =
      this.pocsagMessagesContainer.querySelectorAll(".pocsag-message");

    for (const messageElement of messageElements) {
      const addressInput = messageElement.querySelector(
        '[data-message-field="address"]'
      );
      const messageInput = messageElement.querySelector(
        '[data-message-field="message"]'
      );
      const functionBitsInput = messageElement.querySelector(
        '[data-message-field="functionBits"]'
      );

      const message = {
        address: parseInt(addressInput.value),
        message: messageInput.value.trim(),
      };

      // Add per-message function bits if specified
      if (functionBitsInput && functionBitsInput.value.trim()) {
        message.functionBits = parseInt(functionBitsInput.value);
      }

      args.messages.push(message);
    }

    return args;
  }

  addPOCSAGMessage() {
    const messageCount =
      this.pocsagMessagesContainer.querySelectorAll(".pocsag-message").length;
    const messageIndex = messageCount;

    const messageHtml = `
      <div class="pocsag-message" data-message-index="${messageIndex}">
        <div class="message-header">
          <span>#${messageIndex + 1}</span>
          <button type="button" class="remove-message-btn" title="Remove message">❌</button>
        </div>
        <div class="message-fields">
          <div class="form-group">
            <label for="pocsagAddress${messageIndex}" class="required">Address</label>
            <input
              type="number"
              id="pocsagAddress${messageIndex}"
              min="0"
              placeholder="123456"
              data-message-field="address"
            />
          </div>
          <div class="form-group">
            <label for="pocsagMessage${messageIndex}" class="required">Message</label>
            <textarea
              id="pocsagMessage${messageIndex}"
              placeholder="MESSAGE TEXT"
              rows="2"
              data-message-field="message"
            ></textarea>
          </div>
          <div class="form-group">
            <label for="pocsagMessageFunctionBits${messageIndex}">Function Bits (optional)</label>
            <select id="pocsagMessageFunctionBits${messageIndex}" data-message-field="functionBits">
              <option value="">Use global setting</option>
              <option value="0">0</option>
              <option value="1">1</option>
              <option value="2">2</option>
              <option value="3">3</option>
            </select>
          </div>
        </div>
      </div>
    `;

    this.pocsagMessagesContainer.insertAdjacentHTML("beforeend", messageHtml);
    this.bindPOCSAGMessageEvents();
    if (!this.isRestoring) {
      this.saveState();
    }
    this.validateForm();
  }

  bindPOCSAGMessageEvents() {
    // Bind events for all message fields
    const messageElements =
      this.pocsagMessagesContainer.querySelectorAll(".pocsag-message");

    messageElements.forEach((messageElement) => {
      // Bind input events for validation and state saving
      const inputs = messageElement.querySelectorAll("input, textarea, select");
      inputs.forEach((input) => {
        input.removeEventListener("input", this.onPOCSAGMessageChange);
        input.removeEventListener("change", this.onPOCSAGMessageChange);
        input.addEventListener("input", this.onPOCSAGMessageChange);
        input.addEventListener("change", this.onPOCSAGMessageChange);
      });

      // Bind remove button events
      const removeBtn = messageElement.querySelector(".remove-message-btn");
      if (removeBtn) {
        removeBtn.removeEventListener("click", this.onRemovePOCSAGMessage);
        removeBtn.addEventListener("click", this.onRemovePOCSAGMessage);
      }
    });
  }

  onPOCSAGMessageChange = () => {
    this.saveState();
    this.validateForm();
  };

  onRemovePOCSAGMessage = (event) => {
    const messageElement = event.target.closest(".pocsag-message");
    const messageElements =
      this.pocsagMessagesContainer.querySelectorAll(".pocsag-message");

    // Don't allow removing the last message
    if (messageElements.length <= 1) {
      return;
    }

    messageElement.remove();
    this.renumberPOCSAGMessages();
    this.saveState();
    this.validateForm();
  };

  renumberPOCSAGMessages() {
    const messageElements =
      this.pocsagMessagesContainer.querySelectorAll(".pocsag-message");

    messageElements.forEach((messageElement, index) => {
      messageElement.setAttribute("data-message-index", index);
      messageElement.querySelector("span").textContent = `#${index + 1}`;

      // Update input IDs and labels
      const addressInput = messageElement.querySelector(
        '[data-message-field="address"]'
      );
      const messageInput = messageElement.querySelector(
        '[data-message-field="message"]'
      );
      const functionBitsInput = messageElement.querySelector(
        '[data-message-field="functionBits"]'
      );

      if (addressInput) {
        addressInput.id = `pocsagAddress${index}`;
        messageElement
          .querySelector(`label[for^="pocsagAddress"]`)
          .setAttribute("for", `pocsagAddress${index}`);
      }

      if (messageInput) {
        messageInput.id = `pocsagMessage${index}`;
        messageElement
          .querySelector(`label[for^="pocsagMessage"]`)
          .setAttribute("for", `pocsagMessage${index}`);
      }

      if (functionBitsInput) {
        functionBitsInput.id = `pocsagMessageFunctionBits${index}`;
        messageElement
          .querySelector(`label[for^="pocsagMessageFunctionBits"]`)
          .setAttribute("for", `pocsagMessageFunctionBits${index}`);
      }
    });
  }

  startExecution() {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log("❌ WebSocket not connected", "system");
      return;
    }

    const module = this.moduleSelect.value;
    this.debug("📡 Starting transmission with:", module);
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

      case "pocsag":
        args = this.buildPOCSAGArgs();
        timeout = 0; // No timeout for pocsag by default
        break;

      case "pift8":
        args = {
          frequency: parseFloat(this.ft8FreqInput.value),
          message: this.ft8MessageInput.value.trim(),
        };
        // Add optional fields only if they have values
        if (this.ft8PPMInput.value.trim()) {
          args.ppm = parseFloat(this.ft8PPMInput.value);
        }
        if (this.ft8OffsetInput.value.trim()) {
          args.offset = parseFloat(this.ft8OffsetInput.value);
        }
        if (this.ft8SlotInput.value.trim()) {
          args.slot = parseInt(this.ft8SlotInput.value);
        }
        if (this.ft8RepeatInput.classList.contains("active")) {
          args.repeat = true;
        }
        timeout = 0; // No timeout for ft8 by default
        break;

      case "pisstv":
        args = {
          frequency: parseFloat(this.pisstvFreqInput.value),
          pictureFile: this.pisstvPictureFileInput.value,
        };
        timeout = 0; // No timeout for pisstv by default
        break;

      case "pirtty":
        args = {
          frequency: parseFloat(this.pirttyFreqInput.value),
          message: this.pirttyMessageInput.value.trim(),
        };
        // Add spaceFrequency only if it has a value
        if (this.pirttySpaceFreqInput.value && this.pirttySpaceFreqInput.value.trim() !== "") {
          args.spaceFrequency = parseInt(this.pirttySpaceFreqInput.value);
        }
        timeout = 0; // No timeout for pirtty by default
        break;

      case "fsk":
        args = {
          frequency: parseFloat(this.fskFreqInput.value),
          inputType: this.fskInputTypeText.classList.contains('active') ? "text" : "file",
        };
        if (this.fskInputTypeText.classList.contains('active')) {
          args.text = this.fskTextInput.value.trim();
        } else {
          args.file = this.fskFileInput.value;
        }
        // Add baudRate only if it has a value
        if (this.fskBaudRateInput.value && this.fskBaudRateInput.value.trim() !== "") {
          args.baudRate = parseInt(this.fskBaudRateInput.value);
        }
        timeout = 0; // No timeout for fsk by default
        break;
      case "audiosock-broadcast":
        // For audiosock-broadcast, we first need to connect to /wsunix to get socket path
        this.startAudioSockBroadcast();
        return; // Return early, don't send normal rpitx message yet
    }

    const message = {
      type: "rpitx.execution.start",
      data: {
        moduleName: module,
        args: args,
        timeout: timeout,
        playOnce:
          module === "pifmrds"
            ? this.playModeToggle.classList.contains("active")
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

    // Disconnect unix socket if it exists (for USB AudioSock Broadcast)
    if (this.unixSocket && this.unixSocket.readyState === WebSocket.OPEN) {
      this.log("🔌 Disconnecting from unix socket bridge...", "system");
      this.stopMicrophoneCapture(); // Stop live audio capture
      this.unixSocket.close();
      this.unixSocket = null;
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

    // Clean up unix socket connection if it exists (for USB AudioSock Broadcast)
    if (this.unixSocket && this.unixSocket.readyState === WebSocket.OPEN) {
      this.log("🔌 Closing unix socket after execution stopped", "system");
      this.stopMicrophoneCapture();
      this.unixSocket.close();
      this.unixSocket = null;
    }
  }

  onExecutionError(data) {
    this.setExecutionMode(false);

    this.log(`❌ EXECUTION ERROR: ${data.error}`, "system");
    this.log(`Message: ${data.message}`, "system");

    // Clean up unix socket connection if it exists (for USB AudioSock Broadcast)
    if (this.unixSocket && this.unixSocket.readyState === WebSocket.OPEN) {
      this.log("🔌 Closing unix socket due to execution error", "system");
      this.stopMicrophoneCapture();
      this.unixSocket.close();
      this.unixSocket = null;
    }
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
    formData.append("module", "pifmrds");

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
    formData.append("module", this.moduleSelect.value);

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
      formData.append("module", "pifmrds");

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

  async loadAudioFiles(selectFilename = null) {
    return this.loadFiles({
      endpoint: window.PIrateRFConfig.paths.audioUploadFiles,
      selectElement: this.audioInput,
      fileTypes: ['.wav'],
      selectLatest: !selectFilename,
      selectFilename,
      savedStateKey: 'audio',
      onChangeCallback: () => this.onAudioFileChange(),
      noFilesText: "No audio files",
      debugPrefix: "audio files",
      useServerPath: true,
      pathType: "uploads"
    });
  }

  async loadImageFiles(selectFilename = null) {
    // Load spectrum paint images (.Y files)
    await this.loadFiles({
      endpoint: window.PIrateRFConfig.paths.imageUploadFiles,
      selectElement: this.pictureFileInput,
      fileTypes: ['.Y'],
      selectLatest: !selectFilename,
      selectFilename: selectFilename && selectFilename.endsWith('.Y') ? selectFilename : null,
      savedStateKey: selectFilename ? null : 'spectrumpaint',
      onChangeCallback: () => this.onImageFileChange(),
      noFilesText: "No .Y image files",
      debugPrefix: "spectrum paint images",
      useServerPath: true,
      pathType: "imageUploads"
    });

    // Load PISSTV images (.rgb files)
    await this.loadFiles({
      endpoint: window.PIrateRFConfig.paths.imageUploadFiles,
      selectElement: this.pisstvPictureFileInput,
      fileTypes: ['.rgb'],
      selectLatest: !selectFilename,
      selectFilename: selectFilename && selectFilename.endsWith('.rgb') ? selectFilename : null,
      savedStateKey: selectFilename ? null : 'pisstv',
      onChangeCallback: () => this.onPisstvImageFileChange(),
      noFilesText: "No .rgb image files",
      debugPrefix: "PISSTV images",
      useServerPath: true,
      pathType: "imageUploads"
    });
  }

  onImageFileChange() {
    const hasSelection =
      this.pictureFileInput.value && this.pictureFileInput.value !== "";
    this.editImageBtn.disabled = !hasSelection;
  }

  onPisstvImageFileChange() {
    const hasSelection =
      this.pisstvPictureFileInput.value && this.pisstvPictureFileInput.value !== "";
    this.editPisstvImageBtn.disabled = !hasSelection;
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

    this.currentEditFile = selectedValue.replace(`${serverAudioPath}/`, "");

    // Set the filename in the input
    this.editFileName.value = this.currentEditFile;

    // Show the modal
    this.fileEditModal.style.display = "flex";
  }


  closeEditModal() {
    // Reset all modal state regardless of type
    this.fileEditModal.style.display = "none";
    this.currentEditFile = null;
    this.currentEditDirectory = null;
    this.currentFileExtension = null;
    this.editFileName.value = "";
  }

  renameFile() {
    const newFileName = this.editFileName.value.trim();

    if (!newFileName || !this.currentEditFile) {
      return;
    }

    // Add the extension back to the new filename
    const finalFileName = this.currentFileExtension ? newFileName + this.currentFileExtension : newFileName;

    // Remove extension from current file for comparison
    let currentNameWithoutExtension = this.currentEditFile;
    if (this.currentFileExtension && this.currentEditFile.endsWith(this.currentFileExtension)) {
      currentNameWithoutExtension = this.currentEditFile.slice(0, -this.currentFileExtension.length);
    }

    if (newFileName === currentNameWithoutExtension) {
      // No change, just close
      this.closeEditModal();
      return;
    }

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log("❌ WebSocket not connected", "system");
      return;
    }

    this.showLoadingScreen("Renaming file...");

    // Get the correct file path based on current edit directory
    let filePath;
    if (this.currentEditDirectory === "data") {
      filePath = this.fskFileInput.value;
    } else if (this.currentEditDirectory === "imageUploads") {
      filePath = this.pictureFileInput.value || this.pisstvPictureFileInput.value;
    } else {
      filePath = this.audioInput.value;
    }

    const message = {
      type: "file.rename",
      data: {
        filePath: filePath, // Full path to current file
        newName: finalFileName, // Just the new filename with extension
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
    if (!this.currentEditFile) {
      return;
    }

    if (
      !confirm(`Are you sure you want to delete ${this.currentEditFile}?`)
    ) {
      return;
    }

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log("❌ WebSocket not connected", "system");
      return;
    }

    this.showLoadingScreen("Deleting file...");

    // Get the correct file path based on current edit directory
    let filePath;
    if (this.currentEditDirectory === "data") {
      filePath = this.fskFileInput.value;
    } else if (this.currentEditDirectory === "imageUploads") {
      filePath = this.pictureFileInput.value || this.pisstvPictureFileInput.value;
    } else {
      filePath = this.audioInput.value;
    }

    // Send websocket message for file delete
    const message = {
      type: "file.delete",
      data: {
        filePath: filePath, // Full path to current file
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
    this.loadAudioFiles(data.newName).then(() => {
      this.validateForm();
      this.saveState();
    });
  }

  onFileRenameError(data) {
    this.hideLoadingScreen();
    this.log(`❌ Failed to rename file: ${data.message}`, "system");
    this.showErrorNotification(`Failed to rename file: ${data.message}`, "file-rename");
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
    this.closeEditModal();
    // Refresh the dropdown to show current state
    this.loadAudioFiles();
  }

  // Data file handlers
  onDataFileRenameSuccess(data) {
    this.hideLoadingScreen();
    this.log(
      `✅ Data file renamed from ${data.fileName} to ${data.newName}`,
      "system"
    );
    this.closeEditModal();

    // Refresh the dropdown and select the renamed file
    this.loadDataFiles(data.newName).then(() => {
      this.validateForm();
      this.saveState();
    });
  }

  onDataFileRenameError(data) {
    this.hideLoadingScreen();
    this.log(`❌ Failed to rename data file: ${data.message}`, "system");
  }

  onDataFileDeleteSuccess(data) {
    this.hideLoadingScreen();
    this.log(`✅ Data file deleted: ${data.fileName}`, "system");
    this.closeEditModal();

    // Refresh the dropdown
    this.loadDataFiles();
  }

  onDataFileDeleteError(data) {
    this.hideLoadingScreen();
    this.log(`❌ Failed to delete data file: ${data.message}`, "system");
    this.closeEditModal();
    // Refresh the dropdown to show current state
    this.loadDataFiles();
  }

  // This function is defined earlier (line 1252) with unified modal support


  renameImageFile() {
    const newFileName = this.editFileName.value.trim();

    if (!newFileName || !this.currentEditFile) {
      return;
    }

    // Add the extension back to the new filename
    const finalFileName = this.currentFileExtension ? newFileName + this.currentFileExtension : newFileName;

    // Remove extension from current file for comparison
    let currentNameWithoutExtension = this.currentEditFile;
    if (this.currentFileExtension && this.currentEditFile.endsWith(this.currentFileExtension)) {
      currentNameWithoutExtension = this.currentEditFile.slice(0, -this.currentFileExtension.length);
    }

    if (newFileName === currentNameWithoutExtension) {
      // No change, just close
      this.closeEditModal();
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
        newName: finalFileName, // Just the new filename with extension
      },
    };

    this.ws.send(JSON.stringify(message));
    this.showLoadingScreen();

    // Don't close modal yet - wait for response
  }

  deleteImageFile() {
    if (!this.currentEditFile) {
      return;
    }

    if (
      !confirm(
        `Are you sure you want to delete "${this.currentEditFile}"?`
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
        filePath: this.state.modulename === "pisstv"
          ? this.pisstvPictureFileInput.value
          : this.pictureFileInput.value, // Full path to file (same as audio pattern)
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

    this.closeEditModal();

    // Refresh the dropdown and select the renamed file
    this.loadImageFiles(data.newName).then(() => {
      this.validateForm();
      this.saveState();
    });
  }

  onImageFileRenameError(data) {
    this.hideLoadingScreen();
    this.log(`❌ Failed to rename image file: ${data.message}`, "system");
  }

  onImageFileDeleteSuccess(data) {
    this.debug("onImageFileDeleteSuccess called");
    this.hideLoadingScreen();
    this.log(`✅ Image file deleted: ${data.fileName}`, "system");
    this.closeEditModal();
    this.loadImageFiles();
  }

  onImageFileDeleteError(data) {
    this.hideLoadingScreen();
    this.log(`❌ Failed to delete image file: ${data.message}`, "system");
    this.closeEditModal();
    // Refresh the dropdown to show current state
    this.loadImageFiles();
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
    if (this.playModeToggle.classList.contains("active")) {
      // Switch to continuous mode
      this.playModeToggle.classList.remove("active");
      this.playModeToggle.textContent = "🔁";
      this.playModeToggle.title = "Loop";
    } else {
      // Switch to play once mode
      this.playModeToggle.classList.add("active");
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


  // Unified file loading function for all file types
  async loadFiles(config) {
    const {
      endpoint,
      selectElement,
      fileTypes = [],
      selectLatest = true,
      selectFilename = null,
      savedStateKey,
      onChangeCallback,
      noFilesText = "No files",
      debugPrefix = "files",
      useServerPath = false,
      pathType = "uploads"
    } = config;

    try {
      const response = await this.customFetch(endpoint, {}, `Loading ${debugPrefix}...`);
      if (!response.ok) {
        throw new Error(`Failed to load ${debugPrefix}: ${response.status}`);
      }

      const files = await response.json();

      if (this.isDebugMode) {
        this.log(`🔄 Loaded ${files.length} ${debugPrefix}`, "system");
      }

      // Sort by modTime (newest first)
      files.sort((a, b) => new Date(b.modTime) - new Date(a.modTime));

      // Clear existing options
      selectElement.innerHTML = '';

      // Filter files by type if specified
      let filteredFiles = files.filter(file => !file.isDir);
      if (fileTypes.length > 0) {
        filteredFiles = filteredFiles.filter(file =>
          fileTypes.some(ext => file.name.endsWith(ext))
        );
      }

      if (filteredFiles.length === 0) {
        const option = document.createElement("option");
        option.value = "";
        option.textContent = noFilesText;
        option.disabled = true;
        selectElement.appendChild(option);
      } else {
        // Add files to dropdown
        filteredFiles.forEach(file => {
          const option = document.createElement('option');
          const fileName = typeof file === 'string' ? file : file.name;
          option.value = useServerPath ? this.buildFilePath(fileName, pathType, true) : fileName;
          option.textContent = fileName;
          selectElement.appendChild(option);
        });

        // Handle selection
        if (selectFilename) {
          // Try to select the specific filename
          const targetOption = Array.from(selectElement.options).find(option =>
            option.textContent === selectFilename
          );
          if (targetOption) {
            selectElement.value = targetOption.value;
          } else {
            // If not found, select the first one
            selectElement.selectedIndex = 0;
          }
        } else if (selectLatest) {
          selectElement.selectedIndex = 0;
        } else if (savedStateKey) {
          this.selectSavedOrFirstFile(selectElement, savedStateKey, onChangeCallback);
        }
      }

      // Trigger change callback if provided
      if (onChangeCallback) {
        onChangeCallback();
      }

      this.validateForm();
    } catch (error) {
      this.log(`❌ Failed to load ${debugPrefix}: ${error.message}`, "system");
    }
  }

  // Unified function to select saved file or first file
  selectSavedOrFirstFile(selectElement, savedStateKey, onChangeCallback) {
    let savedValue = null;

    // Handle nested state keys like 'spectrumpaint.pictureFile'
    if (savedStateKey === 'spectrumpaint' && this.state.spectrumpaint) {
      savedValue = this.state.spectrumpaint.pictureFile;
    } else if (savedStateKey === 'pisstv' && this.state.pisstv) {
      savedValue = this.state.pisstv.pictureFile;
    } else if (savedStateKey === 'fsk' && this.state.fsk) {
      savedValue = this.state.fsk.file;
    } else if (savedStateKey === 'audio' && this.state.audio) {
      savedValue = this.state.audio;
    } else if (this.state[savedStateKey]) {
      savedValue = this.state[savedStateKey];
    }

    if (savedValue) {
      for (let i = 0; i < selectElement.options.length; i++) {
        if (selectElement.options[i].value === savedValue) {
          selectElement.selectedIndex = i;
          if (onChangeCallback) onChangeCallback();
          return;
        }
      }
    }

    // Fallback: select first (newest) file if no saved selection or saved file doesn't exist
    if (selectElement.options.length > 0 && !selectElement.options[0].disabled) {
      selectElement.selectedIndex = 0;
      if (onChangeCallback) onChangeCallback();
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
    this.debug("💾 Saving state to localStorage");
    // Ensure state object structure exists
    if (!this.state.pifmrds) this.state.pifmrds = {};
    if (!this.state.morse) this.state.morse = {};
    if (!this.state.tune) this.state.tune = {};
    if (!this.state.spectrumpaint) this.state.spectrumpaint = {};
    if (!this.state.pichirp) this.state.pichirp = {};
    if (!this.state.pocsag) this.state.pocsag = {};
    if (!this.state.pift8) this.state.pift8 = {};
    if (!this.state.pisstv) this.state.pisstv = {};
    if (!this.state.pirtty) this.state.pirtty = {};
    if (!this.state.fsk) this.state.fsk = {};

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
      this.playModeToggle.classList.contains("active");
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

    // Update POCSAG state
    this.state.pocsag.frequency =
      document.getElementById("pocsagFreq")?.value || "";
    this.state.pocsag.baudRate =
      document.getElementById("pocsagBaudRate")?.value || "";
    this.state.pocsag.functionBits =
      document.getElementById("pocsagFunctionBits")?.value || "";
    this.state.pocsag.repeatCount =
      document.getElementById("pocsagRepeatCount")?.value || "";
    this.state.pocsag.numericMode =
      document.getElementById("pocsagNumericMode")?.classList.contains('active') || false;
    this.state.pocsag.invertPolarity =
      document.getElementById("pocsagInvertPolarity")?.classList.contains('active') || false;
    this.state.pocsag.debug =
      document.getElementById("pocsagDebug")?.classList.contains('active') || false;

    // Collect messages state
    this.state.pocsag.messages = [];
    const messageElements = document.querySelectorAll(".pocsag-message");
    messageElements.forEach((messageEl) => {
      const address = messageEl.querySelector('[data-message-field="address"]')?.value || "";
      const message = messageEl.querySelector('[data-message-field="message"]')?.value || "";
      const functionBits = messageEl.querySelector('[data-message-field="functionBits"]')?.value || "";

      if (address || message) {
        this.state.pocsag.messages.push({
          address: address,
          message: message,
          functionBits: functionBits,
        });
      }
    });

    // Update FT8 state
    this.state.pift8.frequency = this.ft8FreqInput.value;
    this.state.pift8.message = this.ft8MessageInput.value;
    this.state.pift8.ppm = this.ft8PPMInput.value;
    this.state.pift8.offset = this.ft8OffsetInput.value;
    this.state.pift8.slot = this.ft8SlotInput.value;
    this.state.pift8.repeat = this.ft8RepeatInput.classList.contains("active");

    // Update PISSTV state
    this.state.pisstv.frequency = this.pisstvFreqInput.value;
    this.state.pisstv.pictureFile = this.pisstvPictureFileInput.value;

    // Update PIRTTY state
    this.state.pirtty.frequency = this.pirttyFreqInput.value;
    this.state.pirtty.spaceFrequency = this.pirttySpaceFreqInput.value;
    this.state.pirtty.message = this.pirttyMessageInput.value;

    // Update FSK state
    this.state.fsk.frequency = this.fskFreqInput.value;
    this.state.fsk.inputType = this.fskInputTypeText.classList.contains('active') ? "text" : "file";
    this.state.fsk.text = this.fskTextInput.value;
    this.state.fsk.file = this.fskFileInput.value;
    this.state.fsk.baudRate = this.fskBaudRateInput.value;

    // Update AudioSock Broadcast state
    this.state["audiosock-broadcast"].frequency = this.audioSockBroadcastFreqInput.value;
    this.state["audiosock-broadcast"].sampleRate = this.audioSockBroadcastSampleRateInput.value;
    this.state["audiosock-broadcast"].bufferSize = this.audioSockBroadcastBufferSizeInput.value;
    this.state["audiosock-broadcast"].modulation = this.audioSockBroadcastModulationInput.value;
    this.state["audiosock-broadcast"].gain = this.audioSockBroadcastGainInput.value;

    this.debug("💾 Saving FSK state:", this.state.fsk);
    this.debug("💾 Saving FT8 state to localStorage, fucking finally:", this.state.pift8);
    if (this.isDebugMode) {
      console.trace("📍 saveState() called from:");
    }

    try {
      localStorage.setItem("piraterf_state", JSON.stringify(this.state));
    } catch (e) {
      console.warn("Failed to save state to localStorage:", e);
    }
  }

  // Restore just the module selection immediately to show correct form
  restoreModuleSelection() {
    try {
      const savedState = localStorage.getItem("piraterf_state");
      if (savedState) {
        const parsedState = JSON.parse(savedState);
        if (parsedState.modulename) {
          // Set flag to prevent saveState during restoration
          this.isRestoring = true;
          this.moduleSelect.value = parsedState.modulename;
          this.onModuleChange();
          this.isRestoring = false;
        }
      }
    } catch (e) {
      console.warn("Failed to restore module selection from localStorage:", e);
    }
  }

  // Load state from localStorage and sync to DOM
  restoreState() {
    this.debug("🔄 Starting state restoration from localStorage");
    this.isRestoring = true;
    try {
      const savedState = localStorage.getItem("piraterf_state");
      if (savedState) {
        this.debug("📦 Found saved state, loading that shit");
        this.debug("💾 Raw saved state:", savedState);
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
          pocsag: { ...this.state.pocsag, ...parsedState.pocsag },
          pift8: { ...this.state.pift8, ...parsedState.pift8 },
          pisstv: { ...this.state.pisstv, ...parsedState.pisstv },
          pirtty: { ...this.state.pirtty, ...parsedState.pirtty },
          fsk: { ...this.state.fsk, ...parsedState.fsk },
        };
      }

      // Sync state to DOM elements
      this.debug("🔄 Syncing state to DOM elements");
      this.syncStateToDOM();
      this.debug("✅ State restoration completed, shit works");
    } catch (e) {
      console.warn("Failed to restore state from localStorage:", e);
    } finally {
      this.isRestoring = false;
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
    if (!this.state.pocsag) this.state.pocsag = {};
    if (!this.state.pisstv) this.state.pisstv = {};
    if (!this.state.pirtty) this.state.pirtty = {};
    if (!this.state.fsk) this.state.fsk = {};

    // Sync module selection
    if (this.state.modulename) this.moduleSelect.value = this.state.modulename;

    // Sync PIFMRDS form inputs
    if (this.state.pifmrds.freq) this.freqInput.value = this.state.pifmrds.freq;
    if (this.state.pifmrds.audio)
      this.audioInput.value = this.state.pifmrds.audio;
    if (this.state.pifmrds.pi) this.piInput.value = this.state.pifmrds.pi;
    if (this.state.pifmrds.ps) this.psInput.value = this.state.pifmrds.ps;
    if (this.state.pifmrds.rt) this.rtInput.value = this.state.pifmrds.rt;
    if (this.state.pifmrds.timeout) {
      const timeoutEl = document.getElementById("timeout");
      if (timeoutEl) timeoutEl.value = this.state.pifmrds.timeout;
    }

    // Sync play mode toggle (PIFMRDS only)
    if (this.state.pifmrds.playOnce !== undefined) {
      if (this.state.pifmrds.playOnce) {
        this.playModeToggle.classList.add("active");
        this.playModeToggle.textContent = "⏭️";
        this.playModeToggle.title = "Play once";
      } else {
        this.playModeToggle.classList.remove("active");
        this.playModeToggle.textContent = "🔁";
        this.playModeToggle.title = "Loop";
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
    if (this.state.morse.freq) {
      const morseFreqEl = document.getElementById("morseFreq");
      if (morseFreqEl) morseFreqEl.value = this.state.morse.freq;
    }
    if (this.state.morse.rate) {
      const morseRateEl = document.getElementById("morseRate");
      if (morseRateEl) morseRateEl.value = this.state.morse.rate;
    }
    if (this.state.morse.message) {
      const morseMessageEl = document.getElementById("morseMessage");
      if (morseMessageEl) morseMessageEl.value = this.state.morse.message;
    }

    // Sync TUNE form inputs
    if (this.state.tune.freq) {
      const tuneFreqEl = document.getElementById("tuneFreq");
      if (tuneFreqEl) tuneFreqEl.value = this.state.tune.freq;
    }
    if (this.state.tune.exitImmediate !== undefined) {
      const tuneExitImmediateEl = document.getElementById("tuneExitImmediate");
      if (tuneExitImmediateEl) tuneExitImmediateEl.checked = this.state.tune.exitImmediate;
    }
    if (this.state.tune.ppm) {
      const tunePPMEl = document.getElementById("tunePPM");
      if (tunePPMEl) tunePPMEl.value = this.state.tune.ppm;
    }

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

    // Sync POCSAG form inputs
    if (this.state.pocsag.frequency) {
      const freqEl = document.getElementById("pocsagFreq");
      if (freqEl) freqEl.value = this.state.pocsag.frequency;
    }
    if (this.state.pocsag.baudRate) {
      const baudRateEl = document.getElementById("pocsagBaudRate");
      if (baudRateEl) baudRateEl.value = this.state.pocsag.baudRate;
    }
    if (this.state.pocsag.functionBits) {
      const functionBitsEl = document.getElementById("pocsagFunctionBits");
      if (functionBitsEl) functionBitsEl.value = this.state.pocsag.functionBits;
    }
    if (this.state.pocsag.repeatCount) {
      const repeatCountEl = document.getElementById("pocsagRepeatCount");
      if (repeatCountEl) repeatCountEl.value = this.state.pocsag.repeatCount;
    }
    if (this.state.pocsag.numericMode !== undefined) {
      const numericModeEl = document.getElementById("pocsagNumericMode");
      if (numericModeEl) {
        if (this.state.pocsag.numericMode) {
          numericModeEl.classList.add('active');
        } else {
          numericModeEl.classList.remove('active');
        }
      }
    }
    if (this.state.pocsag.invertPolarity !== undefined) {
      const invertPolarityEl = document.getElementById("pocsagInvertPolarity");
      if (invertPolarityEl) {
        if (this.state.pocsag.invertPolarity) {
          invertPolarityEl.classList.add('active');
        } else {
          invertPolarityEl.classList.remove('active');
        }
      }
    }
    if (this.state.pocsag.debug !== undefined) {
      const debugEl = document.getElementById("pocsagDebug");
      if (debugEl) {
        if (this.state.pocsag.debug) {
          debugEl.classList.add('active');
        } else {
          debugEl.classList.remove('active');
        }
      }
    }

    // Restore POCSAG messages
    if (this.state.pocsag.messages && this.state.pocsag.messages.length > 0) {
      // Clear existing messages first (keep one empty message)
      const messagesContainer = document.getElementById("pocsagMessages");
      messagesContainer.innerHTML = "";

      // Add all saved messages
      this.state.pocsag.messages.forEach((msg, index) => {
        this.addPOCSAGMessage();
        const messageEl = messagesContainer.querySelector(
          `[data-message-index="${index}"]`
        );
        if (messageEl) {
          const addressInput = messageEl.querySelector('[data-message-field="address"]');
          const messageInput = messageEl.querySelector('[data-message-field="message"]');
          const functionBitsInput = messageEl.querySelector('[data-message-field="functionBits"]');

          if (addressInput) addressInput.value = msg.address || "";
          if (messageInput) messageInput.value = msg.message || "";
          if (functionBitsInput) functionBitsInput.value = msg.functionBits || "";
        }
      });

      // If no messages were restored, ensure we have at least one empty message
      if (messagesContainer.children.length === 0) {
        this.addPOCSAGMessage();
      }
    }

    // Sync FT8 form inputs
    this.debug("🔄 Restoring FT8 state from localStorage:", this.state.pift8);
    if (this.state.pift8.frequency) {
      this.debug("📡 Restoring FT8 frequency:", this.state.pift8.frequency);
      this.ft8FreqInput.value = this.state.pift8.frequency;
    }
    if (this.state.pift8.message) {
      this.debug("💬 Restoring FT8 message:", this.state.pift8.message);
      this.ft8MessageInput.value = this.state.pift8.message;
    }
    if (this.state.pift8.ppm) this.ft8PPMInput.value = this.state.pift8.ppm;
    if (this.state.pift8.offset) this.ft8OffsetInput.value = this.state.pift8.offset;
    if (this.state.pift8.slot !== undefined) this.ft8SlotInput.value = this.state.pift8.slot;
    if (this.state.pift8.repeat !== undefined) {
      if (this.state.pift8.repeat) {
        this.ft8RepeatInput.classList.add("active");
      } else {
        this.ft8RepeatInput.classList.remove("active");
      }
    }

    // Sync PISSTV form inputs
    if (this.state.pisstv.frequency && this.pisstvFreqInput)
      this.pisstvFreqInput.value = this.state.pisstv.frequency;
    if (this.state.pisstv.pictureFile && this.pisstvPictureFileInput)
      this.pisstvPictureFileInput.value = this.state.pisstv.pictureFile;

    // Restore PIRTTY state
    if (this.state.pirtty.frequency && this.pirttyFreqInput)
      this.pirttyFreqInput.value = this.state.pirtty.frequency;
    if (this.state.pirtty.spaceFrequency && this.pirttySpaceFreqInput)
      this.pirttySpaceFreqInput.value = this.state.pirtty.spaceFrequency;
    if (this.state.pirtty.message && this.pirttyMessageInput)
      this.pirttyMessageInput.value = this.state.pirtty.message;

    // Restore FSK state
    if (this.state.fsk.frequency && this.fskFreqInput)
      this.fskFreqInput.value = this.state.fsk.frequency;
    if (this.state.fsk.inputType !== undefined) {
      this.setFskInputType(this.state.fsk.inputType, true); // Skip save during restoration
    }
    if (this.state.fsk.text !== undefined && this.fskTextInput)
      this.fskTextInput.value = this.state.fsk.text;
    if (this.state.fsk.file && this.fskFileInput)
      this.fskFileInput.value = this.state.fsk.file;
    if (this.state.fsk.baudRate && this.fskBaudRateInput)
      this.fskBaudRateInput.value = this.state.fsk.baudRate;

    // Restore AudioSock Broadcast state
    if (this.state["audiosock-broadcast"].frequency && this.audioSockBroadcastFreqInput)
      this.audioSockBroadcastFreqInput.value = this.state["audiosock-broadcast"].frequency;
    if (this.state["audiosock-broadcast"].sampleRate && this.audioSockBroadcastSampleRateInput)
      this.audioSockBroadcastSampleRateInput.value = this.state["audiosock-broadcast"].sampleRate;
    if (this.state["audiosock-broadcast"].bufferSize && this.audioSockBroadcastBufferSizeInput)
      this.audioSockBroadcastBufferSizeInput.value = this.state["audiosock-broadcast"].bufferSize;
    if (this.state["audiosock-broadcast"].modulation && this.audioSockBroadcastModulationInput)
      this.audioSockBroadcastModulationInput.value = this.state["audiosock-broadcast"].modulation;
    if (this.state["audiosock-broadcast"].gain && this.audioSockBroadcastGainInput)
      this.audioSockBroadcastGainInput.value = this.state["audiosock-broadcast"].gain;

    // Note: intro/outro selections are restored by restoreSfxSelections() when SFX files are loaded

    // Trigger module change to show/hide appropriate form fields
    this.onModuleChange();
  }

  startAudioSockBroadcast() {
    // Connect to /wsunix endpoint to get socket path
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/wsunix`;

    this.log("🔌 Connecting to unix socket bridge...", "system");

    const unixWs = new WebSocket(wsUrl);

    unixWs.onopen = async () => {
      this.log("✅ Connected to unix socket bridge", "system");
      // Start microphone capture for live audio streaming
      const micSuccess = await this.startMicrophoneCapture(unixWs);
      if (!micSuccess) {
        // Microphone failed, connection will be closed by startMicrophoneCapture
        return;
      }
    };

    unixWs.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        if (message.type === "wsunixbridge.init") {
          this.handleWSUnixBridgeInit(message.data, unixWs);
        }
      } catch (error) {
        this.log(`❌ Error parsing unix socket message: ${error.message}`, "system");
      }
    };

    unixWs.onerror = (error) => {
      this.log("❌ Unix socket connection error", "system");
      console.error("Unix socket WebSocket error:", error);
      // Clean up on error
      this.stopMicrophoneCapture();
    };

    unixWs.onclose = () => {
      this.log("🔌 Unix socket connection closed", "system");
      // Clean up when connection closes
      this.stopMicrophoneCapture();
    };
  }

  handleWSUnixBridgeInit(initData, unixWs) {
    this.log(`📍 Received socket path: ${initData.writerSocket}`, "system");

    // Check if microphone capture is ready - if not, don't start gorpitx
    if (!this.microphoneReady) {
      this.log(`⚠️ Waiting for microphone before starting transmission...`, "system");
      // Store the data for later use when mic is ready
      this.pendingAudioSockArgs = {
        frequency: parseFloat(this.audioSockBroadcastFreqInput.value),
        socketPath: initData.writerSocket,
        unixWs: unixWs
      };
      if (this.audioSockBroadcastSampleRateInput.value && this.audioSockBroadcastSampleRateInput.value.trim() !== "") {
        this.pendingAudioSockArgs.sampleRate = parseInt(this.audioSockBroadcastSampleRateInput.value);
      }
      if (this.audioSockBroadcastModulationInput.value && this.audioSockBroadcastModulationInput.value.trim() !== "") {
        this.pendingAudioSockArgs.modulation = this.audioSockBroadcastModulationInput.value.trim();
      }
      if (this.audioSockBroadcastGainInput.value && this.audioSockBroadcastGainInput.value.trim() !== "") {
        this.pendingAudioSockArgs.gain = parseFloat(this.audioSockBroadcastGainInput.value);
      }
      return;
    }

    // Microphone is ready, start gorpitx immediately
    this.startGorpitxAudioSockModule(initData, unixWs);
  }

  startGorpitxAudioSockModule(initData, unixWs) {
    const args = {
      frequency: parseFloat(this.audioSockBroadcastFreqInput.value),
      socketPath: initData.writerSocket
    };

    // Add optional sample rate if provided
    if (this.audioSockBroadcastSampleRateInput.value && this.audioSockBroadcastSampleRateInput.value.trim() !== "") {
      args.sampleRate = parseInt(this.audioSockBroadcastSampleRateInput.value);
    }

    // Add optional modulation if provided
    if (this.audioSockBroadcastModulationInput.value && this.audioSockBroadcastModulationInput.value.trim() !== "") {
      args.modulation = this.audioSockBroadcastModulationInput.value.trim();
    }

    // Add optional gain if provided
    if (this.audioSockBroadcastGainInput.value && this.audioSockBroadcastGainInput.value.trim() !== "") {
      args.gain = parseFloat(this.audioSockBroadcastGainInput.value);
    }

    // Now start the gorpitx module through normal WebSocket
    const message = {
      type: "rpitx.execution.start",
      data: {
        moduleName: "audiosock-broadcast",
        args: args,
        timeout: 0, // No timeout - run until stopped
        playOnce: false,
        intro: null,
        outro: null,
      },
      id: this.generateUUID(),
    };

    if (this.isDebugMode) {
      this.log("📤 SENDING: " + JSON.stringify(message, null, 2), "send");
    }

    // Send via the main WebSocket connection
    this.ws.send(JSON.stringify(message));

    // Store reference to unix socket for later cleanup
    this.unixSocket = unixWs;
  }

  async startMicrophoneCapture(unixWs) {
    // Initialize microphone status
    this.microphoneReady = false;

    try {
      this.log("🎤 Starting live microphone capture...", "system");

      // Check for browser compatibility
      if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
        throw new Error("getUserMedia not supported in this browser");
      }

      // Check for HTTPS requirement (getUserMedia requires secure context)
      if (location.protocol !== 'https:' && location.hostname !== 'localhost' && location.hostname !== '127.0.0.1') {
        throw new Error("getUserMedia requires HTTPS or localhost");
      }

      // Get configured sample rate or default to 48000
      const configuredSampleRate = this.audioSockBroadcastSampleRateInput.value ?
        parseInt(this.audioSockBroadcastSampleRateInput.value) : 48000;

      // Request microphone access for live streaming
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          sampleRate: configuredSampleRate,
          channelCount: 1,
          echoCancellation: true,
          noiseSuppression: true
        }
      });

      this.log("✅ Microphone access granted for live streaming", "system");
      return await this.processMicrophoneStream(stream, unixWs);

    } catch (error) {
      this.log(`❌ Microphone capture error: ${error.message}`, "system");

      if (error.message.includes("getUserMedia not supported")) {
        this.log("❌ Browser doesn't support microphone access (try Chrome/Firefox)", "system");
      } else if (error.message.includes("requires HTTPS")) {
        this.log("❌ Microphone access requires HTTPS or localhost", "system");
      } else if (error.name === "NotFoundError") {
        this.log("❌ No microphone found for live streaming", "system");
      } else if (error.name === "NotAllowedError") {
        this.log("❌ Microphone permission denied for live streaming", "system");
      } else if (error.name === "NotSupportedError") {
        this.log("❌ Microphone not supported on this device", "system");
      } else if (error.name === "OverconstrainedError") {
        this.log("❌ Microphone constraints not supported (trying fallback)", "system");
        // Try with basic constraints as fallback
        try {
          const basicStream = await navigator.mediaDevices.getUserMedia({ audio: true });
          this.log("✅ Basic microphone access granted", "system");
          // Continue with basic stream processing...
          return await this.processMicrophoneStream(basicStream, unixWs);
        } catch (fallbackError) {
          this.log(`❌ Fallback microphone access failed: ${fallbackError.message}`, "system");
        }
      }

      // Close the unix socket connection since we can't stream audio
      this.log("🔌 Closing connection - no audio available", "system");
      if (unixWs.readyState === WebSocket.OPEN) {
        unixWs.close();
      }

      return false; // Failure
    }
  }

  async processMicrophoneStream(stream, unixWs) {
    try {
      // Get configured sample rate or default to 48000
      const configuredSampleRate = this.audioSockBroadcastSampleRateInput.value ?
        parseInt(this.audioSockBroadcastSampleRateInput.value) : 48000;

      // Create audio context for PCM processing
      const audioContext = new (window.AudioContext || window.webkitAudioContext)({
        sampleRate: configuredSampleRate
      });

      this.log(`🎵 Using sample rate: ${configuredSampleRate} Hz (actual: ${audioContext.sampleRate} Hz)`, "system");

      const source = audioContext.createMediaStreamSource(stream);

      // Try to use AudioWorkletNode (modern approach)
      let processor;
      try {
        // Define the AudioWorklet processor inline as a data URL
        const workletCode = `
          class PCMProcessor extends AudioWorkletProcessor {
            process(inputs, outputs, parameters) {
              const input = inputs[0];
              if (input && input[0]) {
                // Send PCM data to main thread
                this.port.postMessage({
                  type: 'pcm',
                  data: input[0] // Float32Array
                });
              }
              return true;
            }
          }
          registerProcessor('pcm-processor', PCMProcessor);
        `;

        const workletBlob = new Blob([workletCode], { type: 'application/javascript' });
        const workletUrl = URL.createObjectURL(workletBlob);

        await audioContext.audioWorklet.addModule(workletUrl);
        processor = new AudioWorkletNode(audioContext, 'pcm-processor');

        processor.port.onmessage = (event) => {
          if (event.data.type === 'pcm' && unixWs.readyState === WebSocket.OPEN) {
            const float32Data = event.data.data;

            // Convert float32 (-1.0 to 1.0) to int16 (-32768 to 32767)
            const int16Data = new Int16Array(float32Data.length);
            for (let i = 0; i < float32Data.length; i++) {
              const sample = Math.max(-1, Math.min(1, float32Data[i]));
              int16Data[i] = sample < 0 ? sample * 32768 : sample * 32767;
            }

            // Send raw PCM data to Unix socket
            unixWs.send(int16Data.buffer);
          }
        };

        source.connect(processor);
        processor.connect(audioContext.destination);

        const bufferSize = parseInt(this.audioSockBroadcastBufferSizeInput.value) || 4096;
        this.log(`🎵 AudioWorkletNode initialized (buffer size setting: ${bufferSize} samples)`, "system");

      } catch (workletError) {
        this.log("⚠️ Falling back to ScriptProcessorNode", "system");

        // Fallback to ScriptProcessorNode (deprecated but more compatible)
        const bufferSize = parseInt(this.audioSockBroadcastBufferSizeInput.value) || 4096;
        processor = audioContext.createScriptProcessor(bufferSize, 1, 1);
        this.log(`🎵 Using ScriptProcessorNode with buffer size: ${bufferSize} samples (${bufferSize * 2} bytes)`, "system");

        processor.onaudioprocess = (event) => {
          if (unixWs.readyState === WebSocket.OPEN) {
            // Get the float32 PCM data
            const float32Data = event.inputBuffer.getChannelData(0);

            // Convert float32 (-1.0 to 1.0) to int16 (-32768 to 32767)
            const int16Data = new Int16Array(float32Data.length);
            for (let i = 0; i < float32Data.length; i++) {
              const sample = Math.max(-1, Math.min(1, float32Data[i]));
              int16Data[i] = sample < 0 ? sample * 32768 : sample * 32767;
            }

            // Send raw PCM data to Unix socket
            unixWs.send(int16Data.buffer);
          }
        };

        source.connect(processor);
        processor.connect(audioContext.destination);
      }

      // Store references for cleanup
      this.liveAudioStream = stream;
      this.liveAudioContext = audioContext;
      this.liveAudioProcessor = processor;

      this.log("🎵 Live audio streaming started (48kHz mono PCM)", "system");

      // Mark microphone as ready
      this.microphoneReady = true;

      // If we have pending AudioSock args, start gorpitx now
      if (this.pendingAudioSockArgs) {
        this.log("🚀 Microphone ready, starting transmission...", "system");
        this.startGorpitxAudioSockModule({
          writerSocket: this.pendingAudioSockArgs.socketPath
        }, this.pendingAudioSockArgs.unixWs);
        this.pendingAudioSockArgs = null;
      }

      return true; // Success

    } catch (error) {
      this.log(`❌ Audio processing error: ${error.message}`, "system");

      // Clean up the stream
      if (stream) {
        stream.getTracks().forEach(track => track.stop());
      }

      return false;
    }
  }

  stopMicrophoneCapture() {
    if (this.liveAudioStream) {
      this.liveAudioStream.getTracks().forEach(track => track.stop());
      this.liveAudioStream = null;
      this.log("🔇 Stopped live audio stream", "system");
    }

    if (this.liveAudioProcessor) {
      this.liveAudioProcessor.disconnect();
      this.liveAudioProcessor = null;
    }

    if (this.liveAudioContext) {
      this.liveAudioContext.close();
      this.liveAudioContext = null;
    }

    // Clean up flags and pending args
    this.microphoneReady = false;
    this.pendingAudioSockArgs = null;
  }
}

// Initialize the application
document.addEventListener("DOMContentLoaded", () => {
  new PIrateRFController();
});
