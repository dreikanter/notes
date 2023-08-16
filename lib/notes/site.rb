class Notes::Site
  def pages
    [root_page, feed_page] + note_pages + redirect_pages + tag_pages
  end

  def note_pages
    @note_pages ||= note_files.map { Notes::PageBuilder.new(_1).build }.compact.sort_by(&:published_at).reverse
  end

  def tags
    @tags ||= note_pages.map(&:tags).flatten.uniq
  end

  def related_to(page)
    note_pages.filter { _1.uid != page.uid && _1.tags.intersect?(page.tags) }
  end

  def tagged_pages(tag)
    note_pages.filter { _1.tags.include?(tag) }
  end

  private

  def redirect_pages
    note_pages.map do |page|
      Notes::RedirectPage.new(
        template: "redirect.html",
        layout: nil,
        local_path: "#{page.uid}/index.html",
        public_path: page.uid,
        redirect_to: page.url
      )
    end
  end

  def tag_pages
    tags.map do |tag|
      Notes::TagPage.new(
        template: "tag.html",
        layout: "layout.html",
        local_path: "tags/#{tag}/index.html",
        public_path: "tags/#{tag}",
        tag: tag
      )
    end
  end

  def root_page
    Notes::Page.new(
      template: "index.html",
      layout: "layout.html",
      local_path: File.join(Notes::Configuration.site_root_path, "index.html"),
      public_path: Notes::Configuration.site_root_path
    )
  end

  def feed_page
    Notes::Page.new(
      template: "feed.xml",
      layout: nil,
      local_path: File.join(Notes::Configuration.site_root_path, Notes::Configuration.feed_path),
      public_path: Notes::Configuration.feed_path
    )
  end

  def note_files
    Dir.glob("#{Notes::Configuration.notes_path}/**/*.md")
  end
end
