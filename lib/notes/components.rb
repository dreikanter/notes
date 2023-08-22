class Notes::Components
  class << self
    def tags_list(tags:, current_tag: nil)
      tags.sort.map { tag(tag: _1, highlight: _1 == current_tag) }.join
    end

    def pages_list(pages)
      return if pages.none?

      pages.map do |page|
        <<~HTML.strip
          <li>
            <a href="#{page.public_path}">#{page.title}</a>
            <small class="text-slate-400">#{time_tag(time: page.published_at, format: "%Y/%m/%d")}</small>
          </li>
        HTML
      end.then { "<ul>#{_1.join}</ul>" }
    end

    def time_tag(time:, format:)
      time&.then { "<time datetime=\"#{_1.to_time}\">#{_1.strftime(format)}</time>" }
    end

    private

    def tag(tag:, highlight: false)
      classes = %w[block border rounded-md py-0 px-2 my-1 mr-2 no-underline font-normal]

      classes += if highlight
        %w[bg-slate-500 !text-slate-100 border-slate-600]
      else
        %w[bg-slate-100 !text-slate-800 border-slate-200]
      end

      <<~HTML.strip
        <a href="/tags/#{tag}" class="#{classes.join(" ")}">#{tag}</a>
      HTML
    end
  end
end
