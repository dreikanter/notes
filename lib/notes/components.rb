class Notes::Components
  class << self
    include DOTIW::Methods

    def tags_list(tags:, current_tag: nil)
      tags.map { tag(tag: _1, highlight: _1 == current_tag) }.join
    end

    def pages_list(pages)
      return if pages.none?

      pages.map do |page|
        <<~HTML.strip
          <li>
            <a href="#{page.public_path}">#{page.title}</a>
            <small class="text-slate-400">#{ago_timestamp(page)}</small>
          </li>
        HTML
      end.then { "<ul>#{_1.join}</ul>" }
    end

    def ago_timestamp(page)
      page.published_at&.then { time_tag(time: _1, label: ago(_1)) }
    end

    def page_timestamp(page)
      page.published_at&.then { time_tag(time: _1, label: _1.strftime("%Y/%m/%d")) }
    end

    private

    def time_tag(time:, label:)
      "<time datetime=\"#{time.to_time}\">#{label}</time>"
    end

    def ago(time)
      distance_of_time_in_words(time, Time.now, compact: true, highest_measure_only: true)
    end

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
