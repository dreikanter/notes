class Notes::MarkdownParser
  class << self
    def render(source_content)
      instance.render(source_content)
    end

    private

    def instance
      @markdown ||= Redcarpet::Markdown.new(
        Notes::MarkdownRenderer.new(with_toc_data: true),
        fenced_code_blocks: true,
        autolink: true,
        strikethrough: true,
        space_after_headers: true,
        highlight: true
      )
    end
  end
end
