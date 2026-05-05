;;; yune-mode.el --- major mode for the Yune programming language -*- lexical-binding: t; -*-

(defconst yune--keywords
  '("and" "or" "in" "import" "is"))

(defconst yune--types
  '("Int" "Float" "Bool" "String" "Fn" "List" "Type" "Struct" "Union" "Expression"))

(defconst yune--constants
  `("true" "false"))

(defconst yune--operators
  '("+" "-" "*" "/"))

(defconst yune--font-lock-defaults
  `((
     ;; Comments
     ("//.*$" . font-lock-comment-face)
     ;; String literals
     ("\".*?\"\\|#" . font-lock-string-face)
     ;; Regular keywords
     (,(regexp-opt yune--keywords 'words) . font-lock-keyword-face)
     ;; Keywords that are constants
     (,(regexp-opt yune--constants 'words) . font-lock-constant-face)
     ;; Named constants
     ("\\_<[A-Z][A-Z_]+\\_>" . font-lock-constant-face)
     ;; Function names
     ("\\_<\\([a-zA-Z]+\\)\\(\s*(\\|\s+[a-zA-Z0-9(\[{]\\)" 1 font-lock-function-name-face)
     ;; Numeric literals
     ("\\b[0-9]+\\(?:\\.[0-9]+\\)?\\b" . font-lock-constant-face)
     ;; Variables
     ("\\_<[a-z][A-Za-z0-9]*\\_>" . font-lock-variable-name-face)
     ;; Types
     ("\\_<[A-Z][A-Za-z0-9]+\\_>" . font-lock-type-face)
     )))

(defconst yune--indent-offset 4
  "Indentation width for `yune-mode'.")

(defun yune--indent-line ()
  "Indent current line by `yune-indent-offset' spaces."
  (interactive)
  (let ((indent-level
         (save-excursion
           (forward-line -1)
           (if (bobp)
               0
             (current-indentation)))))
    (indent-line-to indent-level)))

(defun yune--newline-and-indent ()
  "Insert newline and indent to previous line's indentation."
  (interactive)
  (newline)
  (yune--indent-line))

(defun yune--increase-indent ()
  "Increase indentation by `yune--indent-offset'."
  (interactive)
  (indent-line-to
   (+ (current-indentation) yune--indent-offset)))

;;;###autoload
(define-derived-mode yune-mode prog-mode "yune"
  "Major mode for the Yune programming language."
  ;; Syntax highlighting
  (setq-local font-lock-defaults yune--font-lock-defaults)
  (setq-local comment-start "//")
  ;; Indentation
  (setq-local indent-tabs-mode nil)
  (setq-local tab-width 4)
  (setq-local indent-line-function #'yune--indent-line)
  (setq-local tab-width yune--indent-offset)
  ;; Fix string highlighting having precedence over comment highlighting
  (setq-local font-lock-syntactic-face-function
              (lambda (state)
                (when (nth 4 state) ;; inside comment
                  font-lock-comment-face)))
  (local-set-key (kbd "RET") #'yune--newline-and-indent)
  (local-set-key (kbd "TAB") #'yune--increase-indent))

;;;###autoload
(add-to-list 'auto-mode-alist '("\\.un\\'" . yune-mode))

(provide 'yune-mode)
