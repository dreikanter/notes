require_relative "./page"

class Site
  attr_reader :configuration

  def initialize(configuration)
    @configuration = configuration
  end

  def pages
    @pages ||= note_files
      .map { Page.new(source_file: _1, configuration: configuration) }
      .filter(&:public?).sort_by(&:published_at).reverse
  end

  def toc_pages
    pages.reject(&:hide_from_toc?)
  end

  def tags
    @tags ||= pages.map(&:tags).flatten.uniq
  end

  def related_to(page)
    pages.filter { _1.uid != page.uid && _1.tags.intersect?(page.tags) }
  end

  def tagged_pages(tag)
    pages.filter { _1.tags.include?(tag) && !_1.hide_from_toc? }.sort_by { _1.published_at }.reverse
  end

  private

  def note_files
    Dir.glob("#{configuration.notes_path}/**/*.md")
  end
end
