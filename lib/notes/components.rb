class Notes::Components
  class << self
    def tags_list(tags:, current_tag: nil)
      tags.map { tag(tag: _1, highlight: _1 == current_tag) }.join
    end

    def pages_list(pages)
      return if pages.none?

      <<~HTML.strip
        <ul>
          <% pages.each do |page| %>
            <li>
              <a href="<%= page.public_path %>"><%= page.title %></a>
              <small><%= timestamp(page) %></small>
            </li>
          <% end %>
        </ul>
      HTML
    end

    def timestamp(page)
      page.published_at&.then do |published_at|
        <<~HTML.strip
          <time datetime="#{published_at.to_time}">#{published_at.strftime("%Y/%m/%d")}</time>
        HTML
      end
    end

    private

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
  end
end
