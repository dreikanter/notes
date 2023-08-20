class Notes::Components
  class << self
    def tag(tag:, highlight: false)
      classes = %w[block border rounded-md py-0 px-2 my-1 mr-2 no-underline font-normal]

      if highlight
        classes += %w[bg-slate-500 !text-slate-100 border-slate-600]
      else
        classes += %w[bg-slate-100 !text-slate-800 border-slate-200]
      end

      <<~HTML.strip
        <a href="/tags/#{tag}" class="#{classes.join(" ")}">#{tag}</a>
      HTML
    end

    def tags_list(tags:, current_tag: nil)
      tags.map { tag(tag: _1, highlight: _1 == current_tag) }.join
    end
  end
end
